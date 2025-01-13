package main

import (
	"os"

	"docgent-backend/internal/infrastructure/genkit"
)

func NewGenkitDocumentAgentConfig() genkit.DocumentAgentConfig {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		panic("GEMINI_API_KEY environment variable is not set")
	}

	return genkit.DocumentAgentConfig{
		GenerativeModelName: "gemini-1.5-flash-001",
		APIKey:              apiKey,
	}
}
