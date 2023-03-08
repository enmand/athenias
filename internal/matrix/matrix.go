package matrix

import (
	"context"

	mapset "github.com/deckarep/golang-set/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// Client is a Matrix client
type Client struct {
	*mautrix.Client
	log *zerolog.Logger

	opts options
}

// options are the options for the Matrix client
type options struct {
	// Log is the logger to use for logging
	Log *zerolog.Logger

	// channels are the channels to join on startup
	Channels mapset.Set[string]

	// SyncStore is the store to use for the client
	SyncStore mautrix.SyncStore

	// StateStore is the store to use for the client
	StateStore mautrix.StateStore

	// DatabaseDSN is the database DSN to use for the client
	DatabaseDSN string
}

// Option is an option for the Matrix client
type Option func(*options)

// WithLogger sets the logger to use for logging
func WithLogger(log *zerolog.Logger) Option {
	return func(o *options) {
		o.Log = log
	}
}

// WithJoinRooms sets the rooms to join on startup
func WithJoinRooms(channels []string) Option {
	return func(o *options) {
		for _, ch := range channels {
			o.Channels.Add(ch)
		}
	}
}

// WithSyncStore sets the store to use for the client
func WithSyncStore(store mautrix.SyncStore) Option {
	return func(o *options) {
		o.SyncStore = store
	}
}

// WithStateStore sets the store to use for the client
func WithStateStore(store mautrix.StateStore) Option {
	return func(o *options) {
		o.StateStore = store
	}
}

// WithDatabaseDSN sets the database DSN to use for the client
func WithDatabaseDSN(dsn string) Option {
	return func(o *options) {
		o.DatabaseDSN = dsn
	}
}

// NewClient creates a new Matrix client
func NewClient(homeserverURL, username, password string, opts ...Option) (*Client, error) {
	o := &options{
		Channels: mapset.NewSet[string](),
	}
	for _, opt := range opts {
		opt(o)
	}

	uid := id.NewUserID(username, homeserverURL)
	client, err := mautrix.NewClient(homeserverURL, uid, password)
	if err != nil {
		return nil, err
	}

	client.Log = *o.Log
	client.Store = o.SyncStore
	client.StateStore = o.StateStore
	if client.StateStore != nil {
		client.Syncer.(mautrix.ExtensibleSyncer).OnEvent(client.StateStoreSyncHandler)
	}

	ch, err := cryptohelper.NewCryptoHelper(client, []byte("athenais"), o.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	ch.LoginAs = &mautrix.ReqLogin{
		Type: mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{
			Type: mautrix.IdentifierTypeUser,
			User: uid.Localpart(),
		},
		Password: password,
	}

	err = ch.Init()
	if err != nil {
		return nil, err
	}

	client.Crypto = ch
	c := &Client{
		Client: client,
		log:    o.Log,

		opts: *o,
	}

	return c, nil
}

func (c *Client) Channels() []string {
	return c.opts.Channels.ToSlice()
}

func (c *Client) OnEvent(f mautrix.EventHandler) {
	s := c.Syncer.(mautrix.ExtensibleSyncer)
	s.OnEvent(f)
}

func (c *Client) OnEventType(evtType event.Type, f mautrix.EventHandler) {
	s := c.Syncer.(mautrix.ExtensibleSyncer)
	s.OnEventType(evtType, f)
}

func (c *Client) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)

	var errch chan error
	go func(errch chan error) error {
		c.SyncPresence = event.PresenceOnline
		err := c.SyncWithContext(ctx)
		if err != nil {
			errch <- err
		}
		return nil
	}(errch)

	if err := c.ensureRooms(); err != nil {
		cancel(err)
		return errors.Wrap(err, "failed to ensure rooms")
	}

	for {
		select {
		case err := <-errch:
			cancel(err)
			return errors.Wrap(err, "failed to sync")
		}
	}
}

func (c *Client) ensureRooms() error {
	resp, err := c.JoinedRooms()
	if err != nil {
		return err
	}

	joinedRooms := mapset.NewSet[string]()
	for _, room := range resp.JoinedRooms {
		joinedRooms.Add(room.String())
	}

	diff := joinedRooms.Difference(c.opts.Channels)
	if diff.Cardinality() > 0 {
		for _, ch := range diff.ToSlice() {
			c.log.Info().Str("channel", ch).Msg("leaving room")
			_, err := c.LeaveRoom(id.RoomID(ch))
			if err != nil {
				return err
			}
		}
	}

	for _, ch := range c.opts.Channels.ToSlice() {
		if !joinedRooms.Contains(ch) {
			c.log.Info().Str("channel", ch).Msg("joining room")
			_, err := c.JoinRoom(ch, "", nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
