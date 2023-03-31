package openai

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/sashabaranov/go-openai"
)

const (
	// DefaultPrompt is the default prompt to use for the system chat
	DefaultPrompt = `This is a conversation with an AI assistant. The assistant is helpful, creative, clever, and very friendly. You are in a Matrix channel. The audience for this conversation is mostly technical.`

	// DefaultChance is the default chance to respond to a message. 0 = never, 100 = always
	DefaultChance = 50
)

// Client is an OpenAI client
type Client struct {
	*openai.Client
	sysPrompt string

	log *zerolog.Logger
}

type options struct {
	prompt string
	log    *zerolog.Logger
}

// Option is an option for the OpenAI client
type Option func(*options)

// WithPrompt sets the prompt to use for the client
func WithPrompt(prompt string) Option {
	return func(o *options) {
		o.prompt = prompt
	}
}

// WithLogger sets the logger to use for logging
func WithLogger(log *zerolog.Logger) Option {
	return func(o *options) {
		o.log = log
	}
}

// NewClient creates a new OpenAI client
func NewClient(apiKey string, opts ...Option) *Client {
	o := &options{}
	o.prompt = DefaultPrompt
	for _, opt := range opts {
		opt(o)
	}

	if o.log == nil {
		l := zerolog.New(nil)
		o.log = &l
	}

	client := openai.NewClient(apiKey)

	return &Client{
		Client:    client,
		sysPrompt: o.prompt,
		log:       o.log,
	}
}

func (c *Client) Prompt(ctx context.Context, prompt string) (string, error) {
	c.log.Debug().Str("prompt", prompt).Str("sys_prompt", c.sysPrompt).Msg("prompting")
	err := c.moderate(ctx, prompt)
	if err != nil {
		return "", err
	}

	resp, err := c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 100,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf(c.sysPrompt),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *Client) moderate(ctx context.Context, input string) error {
	mods, err := c.Moderations(ctx, openai.ModerationRequest{
		Input: input,
	})
	if err != nil {
		return err
	}

	for _, mod := range mods.Results {
		if mod.Flagged {
			return fmt.Errorf("moderation flagged input")
		}
	}

	return nil
}
