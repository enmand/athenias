package sayhi

import (
	"github.com/rs/zerolog"
	"github.com/unerror/athenais/pkg/athenais"
	"maunium.net/go/mautrix/event"
)

type Plugin struct {
	// bot is the bot instance
	bot *athenais.Bot

	// log is the logger to use for logging
	log *zerolog.Logger
}

// NewPlugin creates a new SayHi plugin
func NewPlugin() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "sayhi"
}

func (p *Plugin) Init(bot *athenais.Bot, log *zerolog.Logger) {
	p.log = log
	p.bot = bot

	p.log.Info().Msg("Initializing SayHi plugin")

	bot.Route(
		athenais.Route{
			Handler:   p.handleMessage,
			EventType: event.EventMessage,
		},
	)
}

func (p *Plugin) handleMessage(evt *event.Event) error {
	msg := evt.Content.AsMessage()

	p.log.Debug().Str("msg", msg.Body).Msg("Received message")

	if msg.MsgType == event.MsgText {
		if msg.Body == "!say" {
			p.bot.SendText(evt.RoomID, "Hello!")
		}
	}

	return nil
}
