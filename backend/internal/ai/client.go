package ai

import (
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const systemPrompt = `You are a portfolio analyst for FlowPilot Finance. You have full access to the user's portfolio data. Help them understand their holdings, analyze risk, suggest actions, and discuss strategies.`

const maxConversationMessages = 20

type Client struct {
	client        *anthropic.Client
	conversations map[string][]anthropic.MessageParam
	mu            sync.RWMutex
}

func NewClient(apiKey string) *Client {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Client{
		client:        &client,
		conversations: make(map[string][]anthropic.MessageParam),
	}
}

func (c *Client) GetClient() *anthropic.Client {
	return c.client
}
