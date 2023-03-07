package matrix

import (
	"context"

	_ "github.com/mattn/go-sqlite3"
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
	Channels []id.RoomID

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
			o.Channels = append(o.Channels, id.RoomID(ch))
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
	o := &options{}
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

func (c *Client) OnEvent(evtType event.Type, f mautrix.EventHandler) {
	s := c.Syncer.(mautrix.ExtensibleSyncer)
	s.OnEventType(evtType, f)
}

func (c *Client) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)

	var errch chan error
	go func(errch chan error) error {
		err := c.SyncWithContext(ctx)
		if err != nil {
			errch <- err
		}
		return nil
	}(errch)

	for _, ch := range c.opts.Channels {
		c.JoinRoomByID(ch)
	}

	for {
		select {
		case err := <-errch:
			cancel(err)
			return err
		}
	}
}
