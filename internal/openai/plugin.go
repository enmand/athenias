package openai

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/unerror/athenais/internal/athenais"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

type Plugin struct {
	// client is the OpenAI client to use for the plugin
	client *Client

	// log is the logger to use for logging
	log *zerolog.Logger
}

// NewPlugin creates a new OpenAI plugin
func NewPlugin(aiClient *Client, prompt string, log *zerolog.Logger) *Plugin {
	return &Plugin{
		client: aiClient,
		log:    log,
	}
}

func (p *Plugin) Name() string {
	return "openai"
}

func (p *Plugin) Init(bot *athenais.Bot) error {
	p.log.Info().Msg("Initializing OpenAI plugin")
	bot.OnEvent(event.EventMessage, func(src mautrix.EventSource, evt *event.Event) {
		msg := evt.Content.AsMessage()
		p.log.Info().Interface("msg", msg).Interface("evt", evt).Msg("Received message")
		if evt.Sender == bot.ID() {
			p.log.Debug().Msg("Ignoring message from self")
			return
		}

		if msg.MsgType == event.MsgText {
			out, err := p.client.Prompt(context.Background(), evt.Content.AsMessage().Body)
			if err != nil {
				p.log.Error().Err(err).Msg("Failed to generate response")
				return
			}

			bot.SendText(evt.RoomID, out)
		}

	})

	return nil
}
