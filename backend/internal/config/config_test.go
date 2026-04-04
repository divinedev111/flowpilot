package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_RequiredFields(t *testing.T) {
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("MONGODB_URI")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "mongodb://localhost:27017", cfg.MongoURI)
	assert.Equal(t, "test-key", cfg.AnthropicAPIKey)
	assert.Equal(t, ":8080", cfg.Port)
}

func TestLoad_MissingMongoURI(t *testing.T) {
	os.Unsetenv("MONGODB_URI")
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	_, err := Load()
	assert.Error(t, err)
}

func TestLoad_CustomPort(t *testing.T) {
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	os.Setenv("PORT", "3000")
	defer os.Unsetenv("MONGODB_URI")
	defer os.Unsetenv("ANTHROPIC_API_KEY")
	defer os.Unsetenv("PORT")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, ":3000", cfg.Port)
}
