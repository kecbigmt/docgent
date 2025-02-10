package genai

import (
	"context"
	"docgent/internal/domain"
	"encoding/json"
	"fmt"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ChatModelParams struct {
	fx.In

	Logger *zap.Logger
	Config Config
}

type ChatModel struct {
	logger *zap.Logger
	client *genai.Client
	config Config
}

func NewChatModel(params ChatModelParams) (domain.ChatModel, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, params.Config.ProjectID, params.Config.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return &ChatModel{
		logger: params.Logger,
		client: client,
		config: params.Config,
	}, nil
}

func (c *ChatModel) StartChat(systemInstruction string) domain.ChatSession {
	model := c.client.GenerativeModel(c.config.ModelName)
	model.SystemInstruction = genai.NewUserContent(genai.Text(systemInstruction))
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"toolUse": {Type: genai.TypeString},
		},
		Required: []string{"toolUse"},
	}

	temp := float32(0.1)
	topP := float32(0.5)
	topK := int32(20)
	model.Temperature = &temp
	model.TopP = &topP
	model.TopK = &topK

	session := model.StartChat()
	c.logger.Debug("created chat session", zap.String("model", c.config.ModelName), zap.String("system_instruction", systemInstruction))

	return &ChatSession{
		logger: c.logger,
		model:  model,
		chat:   session,
	}
}

type ChatSession struct {
	logger *zap.Logger
	model  *genai.GenerativeModel
	chat   *genai.ChatSession
}

func (s *ChatSession) SendMessage(ctx context.Context, message string) (string, error) {
	s.chat.History = append(s.chat.History, &genai.Content{
		Role:  "user",
		Parts: []genai.Part{genai.Text(message)},
	})

	s.logger.Debug("sending message", zap.String("role", "user"), zap.String("content", message))

	// Send message
	resp, err := s.chat.SendMessage(ctx, genai.Text(message))
	if err != nil {
		s.logger.Debug("failed to send message", zap.Error(err))
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	// Get response
	var resContent string
	for _, candidate := range resp.Candidates {
		for _, part := range candidate.Content.Parts {
			if text, ok := part.(genai.Text); ok {
				resContent += string(text)
			}
		}
	}

	// Parse JSON response
	response := &struct {
		ToolUse string `json:"toolUse"`
	}{}

	if err := json.Unmarshal([]byte(resContent), response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	s.logger.Debug("received response", zap.String("role", "agent"), zap.String("content", response.ToolUse))

	// Update conversation history
	s.chat.History = append(s.chat.History, &genai.Content{
		Role:  "model",
		Parts: []genai.Part{genai.Text(response.ToolUse)},
	})

	return response.ToolUse, nil
}

func (s *ChatSession) GetHistory() ([]domain.Message, error) {
	history := make([]domain.Message, len(s.chat.History))
	for i, content := range s.chat.History {
		role := domain.UserRole
		if content.Role == "model" {
			role = domain.AssistantRole
		}
		var contentString strings.Builder
		for _, part := range content.Parts {
			if text, ok := part.(genai.Text); ok {
				contentString.WriteString(string(text))
			}
		}
		history[i] = domain.Message{
			Role:    role,
			Content: contentString.String(),
		}
	}
	return history, nil
}
