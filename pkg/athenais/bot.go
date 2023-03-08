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
	r  *Router

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
		r:  NewRouter(),

		log: o.log,
	}

	for _, plug := range o.plugins {
		l := o.log.With().Str("plugin", plug.Name()).Logger()
		plug.Init(b, &l)
	}

	return b
}

// ID returns the UserID of the bot
func (b *Bot) ID() id.UserID {
	return b.mc.UserID
}

// Run runs the bot
func (b *Bot) Run(ctx context.Context) error {
	b.mc.OnEvent(func(_ mautrix.EventSource, evt *event.Event) {
		b.r.Handle(evt)
	})

	return b.mc.Start(ctx)
}

// Route registers a route handler
func (b *Bot) Route(route Route) {
	b.r.AddRoute(route)
}

// SendText sends a text message to a room
func (b *Bot) SendText(roomID id.RoomID, text string) error {
	_, err := b.mc.SendText(roomID, text)
	return err
}
