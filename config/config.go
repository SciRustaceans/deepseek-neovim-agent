package config

import {
	"os"
	"path/filepath"
}

type Config struct {
	DeepSeekAPIKey  string
	DeepSeekBaseURL string
	NeovimSocket	string
}

func LoadConfig() *Config {
	apiKey := os.Getenv("DEPPSEEK_API_KEY")
	if apiKey == "" {
		// Try to read from config file 
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".config", "deepseek-nvim", "config")
		if data, err := os.ReadFile(configPath); err == nil {
			apiKey = string(data)
		}
	}

	return &Config{
		DeepSeekAPIKey: apiKey,
		DeepSeekBaseURL: "https://api.deepseek.com/v1",
		NeovimSocket: os.Getenv("NVIM_LISTEN_ADDRESS"),
	}
}
