package main

import (
	"context"
	"fmt"
	"os"

	"docgent-backend/internal/application/port"
	"docgent-backend/internal/infrastructure/google/vertexai/genai"
	"docgent-backend/internal/infrastructure/google/vertexai/rag"
	raglib "docgent-backend/internal/infrastructure/google/vertexai/rag/lib"

	"golang.org/x/oauth2/google"
)

func newGenAIConfig() genai.Config {
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

	return genai.Config{
		ProjectID: projectID,
		Location:  location,
		ModelName: modelName,
	}
}

func newRAGService() port.RAGService {
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	if projectID == "" {
		panic("GOOGLE_PROJECT_ID environment variable is not set")
	}

	location := os.Getenv("GOOGLE_LOCATION")
	if location == "" {
		panic("GOOGLE_LOCATION environment variable is not set")
	}

	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		panic(fmt.Sprintf("Failed to find default credentials: %v", err))
	}

	return rag.NewService(raglib.NewClientWithCredentials(creds, projectID, location))
}
