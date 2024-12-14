package config

import (
	"errors"
	"os"
)

type Config struct {
	BigIPHost     string
	BigIPUsername string
	BigIPPassword string
	OpenAIKey     string
}

func LoadConfig() (*Config, error) {
	bigipHost := os.Getenv("BIGIP_HOST")
	bigipUser := os.Getenv("BIGIP_USERNAME")
	bigipPass := os.Getenv("BIGIP_PASSWORD")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	if bigipHost == "" || bigipUser == "" || bigipPass == "" || openaiKey == "" {
		return nil, errors.New("missing required environment variables")
	}

	return &Config{
		BigIPHost:     bigipHost,
		BigIPUsername: bigipUser,
		BigIPPassword: bigipPass,
		OpenAIKey:     openaiKey,
	}, nil
}
