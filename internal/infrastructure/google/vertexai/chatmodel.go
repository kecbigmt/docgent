package vertexai

import (
	"context"
	"docgent-backend/internal/domain"
	"encoding/json"
	"fmt"

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
	logger  *zap.Logger
	client  *genai.Client
	model   *genai.GenerativeModel
	history []domain.Message
}

func NewChatModel(params ChatModelParams) (domain.ChatModel, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, params.Config.ProjectID, params.Config.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel(params.Config.ModelName)
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

	return &ChatModel{
		logger:  params.Logger,
		client:  client,
		model:   model,
		history: []domain.Message{},
	}, nil
}

func (c *ChatModel) SetSystemInstruction(instruction string) error {
	c.model.SystemInstruction = genai.NewUserContent(genai.Text(instruction))
	c.logger.Debug("set system instruction", zap.String("instruction", instruction))
	return nil
}

func (c *ChatModel) SendMessage(ctx context.Context, message domain.Message) (string, error) {
	chat := c.model.StartChat()

	// Set up conversation history
	for _, msg := range c.history {
		content := &genai.Content{
			Parts: []genai.Part{genai.Text(msg.Content)},
		}
		if msg.Role == domain.UserRole {
			content.Role = "user"
		} else {
			content.Role = "model"
		}
		chat.History = append(chat.History, content)
	}

	c.logger.Debug("sending message", zap.String("role", "user"), zap.String("content", message.Content))

	// Send message
	resp, err := chat.SendMessage(ctx, genai.Text(message.Content))
	if err != nil {
		c.logger.Debug("failed to send message", zap.Error(err))
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

	c.logger.Debug("received response", zap.String("role", "agent"), zap.String("content", response.ToolUse))

	// Update conversation history
	c.history = append(c.history, message)
	c.history = append(c.history, domain.Message{
		Role:    domain.AssistantRole,
		Content: response.ToolUse,
	})

	return response.ToolUse, nil
}

func (c *ChatModel) GetHistory() ([]domain.Message, error) {
	return c.history, nil
}
