package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port            string
	MongoURI        string
	MongoDB         string
	AnthropicAPIKey string

	SchwabClientID     string
	SchwabClientSecret string
	SchwabRedirectURI  string

	FrontendPort string

	CoinbaseAPIKey    string
	CoinbaseAPISecret string

	EncryptionKey string

	SMTPHost       string
	SMTPPort       string
	SMTPUser       string
	SMTPPass       string
	AlertEmailTo   string
	TelegramToken  string
	TelegramChatID string
	DiscordWebhook string
}

func Load() (*Config, error) {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		return nil, fmt.Errorf("MONGODB_URI is required")
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port

	dbName := os.Getenv("MONGODB_NAME")
	if dbName == "" {
		dbName = "flowpilot"
	}

	return &Config{
		Port:               port,
		MongoURI:           mongoURI,
		MongoDB:            dbName,
		AnthropicAPIKey:    apiKey,
		FrontendPort:       os.Getenv("FRONTEND_PORT"),
		SchwabClientID:     os.Getenv("SCHWAB_CLIENT_ID"),
		SchwabClientSecret: os.Getenv("SCHWAB_CLIENT_SECRET"),
		SchwabRedirectURI:  os.Getenv("SCHWAB_REDIRECT_URI"),
		CoinbaseAPIKey:     os.Getenv("COINBASE_API_KEY"),
		CoinbaseAPISecret:  os.Getenv("COINBASE_API_SECRET"),
		SMTPHost:           os.Getenv("SMTP_HOST"),
		SMTPPort:           os.Getenv("SMTP_PORT"),
		SMTPUser:           os.Getenv("SMTP_USER"),
		SMTPPass:           os.Getenv("SMTP_PASS"),
		AlertEmailTo:       os.Getenv("ALERT_EMAIL_TO"),
		TelegramToken:      os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:     os.Getenv("TELEGRAM_CHAT_ID"),
		DiscordWebhook:     os.Getenv("DISCORD_WEBHOOK_URL"),
		EncryptionKey:      os.Getenv("ENCRYPTION_KEY"),
	}, nil
}
