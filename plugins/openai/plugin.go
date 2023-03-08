package openai

import (
	"context"
	"math/rand"
	"time"

	"github.com/rs/zerolog"
	"github.com/unerror/athenais/pkg/athenais"
	"maunium.net/go/mautrix/event"
)

type Plugin struct {
	// client is the OpenAI client to use for the plugin
	client *Client

	// bot is the bot instance
	bot *athenais.Bot

	// log is the logger to use for logging
	log *zerolog.Logger
}

// NewPlugin creates a new OpenAI plugin
func NewPlugin(aiClient *Client, prompt string) *Plugin {
	rand.Seed(time.Now().UnixNano())

	return &Plugin{
		client: aiClient,
	}
}

func (p *Plugin) Name() string {
	return "openai"
}

func (p *Plugin) Init(bot *athenais.Bot, log *zerolog.Logger) {
	p.log = log
	p.bot = bot

	p.log.Info().Msg("Initializing OpenAI plugin")

	bot.Route(
		athenais.Route{
			Handler:   p.handleMessage,
			EventType: event.EventMessage,
		},
	)
}

func (p *Plugin) handleMessage(evt *event.Event) {
	msg := evt.Content.AsMessage()
	p.log.Info().Interface("msg", msg).Interface("evt", evt).Msg("Received message")

	if evt.Sender == p.bot.ID() {
		p.log.Debug().Msg("Ignoring message from self")
		return
	}

	if msg.MsgType == event.MsgText {
		r := rand.Int() % 100
		p.log.Info().Int("r", r).Msg("Random number")
		if r < 101 {
			out, err := p.client.Prompt(context.Background(), evt.Content.AsMessage().Body)
			if err != nil {
				p.log.Error().Err(err).Msg("Failed to generate response")
				return
			}
			p.bot.SendText(evt.RoomID, out)
		}
	}
}
