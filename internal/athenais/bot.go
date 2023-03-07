package athenais

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/unerror/athenais/internal/matrix"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type options struct {
	log *zerolog.Logger

	plugins []Plugin
}

type Option func(*options)

func WithLogger(log *zerolog.Logger) Option {
	return func(o *options) {
		o.log = log
	}
}

func WithPlugins(plugins ...Plugin) Option {
	return func(o *options) {
		o.plugins = plugins
	}
}

// Bot represents the instance of the bot
type Bot struct {
	mc *matrix.Client

	log *zerolog.Logger
}

// New creates a new instance of the bot
func New(mc *matrix.Client, opts ...Option) *Bot {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	b := &Bot{
		mc: mc,

		log: o.log,
	}

	for _, plug := range o.plugins {
		plug.Init(b)
	}

	return b
}

// ID returns the UserID of the bot
func (b *Bot) ID() id.UserID {
	return b.mc.UserID
}

// Run runs the bot
func (b *Bot) Run(ctx context.Context) error {
	return b.mc.Start(ctx)
}

// OnEvent registers a handler for an event
func (b *Bot) OnEvent(evt event.Type, handler mautrix.EventHandler) {
	b.mc.OnEvent(evt, handler)
}

// SendText sends a text message to a room
func (b *Bot) SendText(roomID id.RoomID, text string) error {
	_, err := b.mc.SendText(roomID, text)
	return err
}
