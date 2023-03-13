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
	log zerolog.Logger

	opts options
}

// options are the options for the Matrix client
type options struct {
	// Log is the logger to use for logging
	Log zerolog.Logger

	// channels are the channels to join on startup
	Channels mapset.Set[string]

	// DatabaseDSN is the database DSN to use for the client
	DatabaseDSN string

	// PickleKey is the crypto pickleKey to use for the client crypto helper
	PickleKey string

	// filter is the filter to use for the client
	Filter *mautrix.Filter

	// client is the client to use for the client
	client *mautrix.Client

	// syncStoreOpts are the options for the sync store
	syncStoreOpts syncOpts

	// stateStoreOpts are the options for the state store
	stateStoreOpts stateOpts

	// chOpts are the options for the crypto helper store
	chStoreOpts chStoreOpts
}

// ClientOption is an option for the Matrix client
type ClientOption func(*options)

// WithLogger sets the logger to use for logging
func WithLogger(log zerolog.Logger) ClientOption {
	return func(o *options) {
		o.Log = log
	}
}

// WithJoinRooms sets the rooms to join on startup
func WithJoinRooms(channels []string) ClientOption {
	return func(o *options) {
		for _, ch := range channels {
			o.Channels.Add(ch)
		}
	}
}

// WithSyncFilter sets the filter to use for the client
func WithSyncFilter(filter *mautrix.Filter) ClientOption {
	return func(o *options) {
		o.Filter = &mautrix.Filter{}
		o.Filter.EventFields = filter.EventFields
		o.Filter.EventFormat = filter.EventFormat
		o.Filter.Presence = filter.Presence
		o.Filter.Room = filter.Room
	}
}

// WithDatabaseDSN sets the database DSN to use for the client
func WithDatabaseDSN(dsn string) ClientOption {
	return func(o *options) {
		o.DatabaseDSN = dsn
	}
}

// WithPickleKey sets the crypto pickleKey to use for the client crypto helper
func WithPickleKey(key string) ClientOption {
	return func(o *options) {
		o.PickleKey = key
	}
}

// WithSyncStore sets the store to use for the client
func WithSyncStore[T syncOpts](opts ...SyncStoreOption[T]) ClientOption {
	return func(o *options) {
		store := new(T)

		for _, opt := range opts {
			opt(store)
		}

		o.syncStoreOpts = *store
	}
}

// WithStateStore sets the store to use for the client
func WithStateStore[T stateOpts](opts ...StateStoreOption[T]) ClientOption {
	return func(o *options) {
		store := new(T)

		for _, opt := range opts {
			opt(store)
		}

		o.stateStoreOpts = *store
	}
}

// WithCryptoHelperStore sets the store to use for the client crypto helper
func WithCryptoHelperStore[T chStoreOpts](opts ...CryptoHelperStoreOption[T]) ClientOption {
	return func(o *options) {
		store := new(T)

		for _, opt := range opts {
			opt(store)
		}

		o.chStoreOpts = *store
	}
}

// NewClient creates a new Matrix client
func NewClient(homeserverURL, username, password string, opts ...ClientOption) (*Client, error) {
	o := &options{
		Channels: mapset.NewSet[string](),
		Filter:   &mautrix.Filter{},
	}

	uid := id.NewUserID(username, homeserverURL)
	client, err := mautrix.NewClient(homeserverURL, uid, password)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(o)
	}

	client.Log = o.Log
	client.Syncer.(*mautrix.DefaultSyncer).FilterJSON = o.Filter

	if o.syncStoreOpts != nil {
		if err := o.syncStoreOpts.Configure(client); err != nil {
			return nil, errors.Wrap(err, "failed to configure sync store")
		}
	}

	if o.stateStoreOpts != nil {
		if err := o.stateStoreOpts.Configure(client); err != nil {
			return nil, errors.Wrap(err, "failed to configure state store")
		}
	}

	if client.StateStore != nil {
		client.Syncer.(mautrix.ExtensibleSyncer).OnEvent(client.StateStoreSyncHandler)
	}

	lreq := &mautrix.ReqLogin{
		Type: mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{
			Type: mautrix.IdentifierTypeUser,
			User: uid.Localpart(),
		},
		Password:         password,
		StoreCredentials: true,
	}

	if o.chStoreOpts != nil {
		ch, err := cryptohelper.NewCryptoHelper(client, []byte(o.PickleKey), o.chStoreOpts.Get())
		if err != nil {
			return nil, err
		}

		if o.chStoreOpts.Managed() {
			ch.LoginAs = lreq
		} else {
			_, err := client.Login(lreq)
			if err != nil {
				return nil, errors.Wrap(err, "failed to login")
			}
		}

		err = ch.Init()
		if err != nil {
			return nil, errors.Wrap(err, "failed to init crypto helper")
		}

		client.Crypto = ch
	} else {
		_, err := client.Login(lreq)
		if err != nil {
			return nil, errors.Wrap(err, "failed to login")
		}
	}

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

	// TODO: include "runtime" rooms that the bot is invited into -- store in DB if not already in events as state?
	if err := c.ensureRooms(); err != nil {
		cancel(err)
		return errors.Wrap(err, "failed to ensure rooms")
	}

	var errch chan error
	go func(errch chan error) error {
		c.SyncPresence = event.PresenceOnline
		err := c.SyncWithContext(ctx)
		if err != nil {
			errch <- err
		}
		return nil
	}(errch)

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
