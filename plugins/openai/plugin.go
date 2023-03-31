package openai

import (
	"context"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/unerror/athenais/pkg/athenais"
	"maunium.net/go/mautrix/event"
)

var (
	openaiKey string
)

type Plugin struct {
	// client is the OpenAI client to use for the plugin
	client *Client

	// bot is the bot instance
	bot *athenais.Bot

	// log is the logger to use for logging
	log *zerolog.Logger

	r *rand.Rand

	cfg *Configuration
}

type Configuration struct {
	// Prompt is the system Prompt to use to prime responses
	Prompt string

	// Chance is the Chance to respond to a message
	Chance int

	// APIKey is the API key for OpenAI
	APIKey string
}

// NewPlugin creates a new OpenAI plugin
func NewPlugin(cfg Configuration) *Plugin {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	aiClient := NewClient(
		cfg.APIKey,
		WithPrompt(cfg.Prompt),
	)

	return &Plugin{
		client: aiClient,
		r:      r,
		cfg:    &cfg,
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

func (p *Plugin) handleMessage(evt *event.Event) error {
	msg := evt.Content.AsMessage()

	if msg.MsgType == event.MsgText {
		r := p.r.Int() % 100
		p.log.Info().Int("r", r).Msg("Random number")
		if r < p.cfg.Chance {
			p.log.Debug().Str("msg", msg.Body).Msg("Responding to message")
			out, err := p.client.Prompt(context.Background(), msg.Body)
			if err != nil {
				p.log.Error().Err(err).Msg("Failed to generate response")
				return errors.Wrap(err, "failed to generate response")
			}

			if err := p.bot.SendText(evt.RoomID, out); err != nil {
				return errors.Wrap(err, "failed to send message")
			}
		}
	}

	return nil
}
