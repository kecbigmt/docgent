package main

import (
	"os"

	"docgent-backend/internal/infrastructure/genkit"
)

func NewGenkitConfig() genkit.Config {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		panic("GEMINI_API_KEY environment variable is not set")
	}

	return genkit.Config{
		GenerativeModelName: "gemini-1.5-flash-001",
		APIKey:              apiKey,
	}
}
