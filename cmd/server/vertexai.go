package main

import (
	"os"

	"docgent-backend/internal/infrastructure/google/vertexai"
)

func NewVertexAIConfig() vertexai.Config {
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	if projectID == "" {
		panic("GOOGLE_PROJECT_ID environment variable is not set")
	}

	location := os.Getenv("GOOGLE_LOCATION")
	if location == "" {
		panic("GOOGLE_LOCATION environment variable is not set")
	}

	modelName := os.Getenv("GOOGLE_MODEL_NAME")
	if modelName == "" {
		panic("GOOGLE_MODEL_NAME environment variable is not set")
	}

	return vertexai.Config{
		ProjectID: projectID,
		Location:  location,
		ModelName: modelName,
	}
}
