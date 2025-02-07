package genkit

import (
	"context"
	"docgent-backend/internal/domain"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type AutoAgentParams struct {
	fx.In

	Logger *zap.Logger
	Config Config
}

type AutoAgent struct {
	logger  *zap.Logger
	model   *genai.GenerativeModel
	history []domain.Message
}

func NewAutoAgent(params AutoAgentParams) (domain.ChatModel, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(params.Config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel(params.Config.GenerativeModelName)
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"toolUse": {Type: genai.TypeString},
		},
		Required: []string{"toolUse"},
	}

	model.SetTemperature(0.1)
	model.SetTopP(0.5)
	model.SetTopK(20)

	return &AutoAgent{
		logger:  params.Logger,
		model:   model,
		history: []domain.Message{},
	}, nil
}

func (a *AutoAgent) SetSystemInstruction(instruction string) error {
	a.model.SystemInstruction = genai.NewUserContent(genai.Text(instruction))
	a.logger.Debug("set system instruction", zap.String("instruction", instruction))
	return nil
}

func (a *AutoAgent) SendMessage(ctx context.Context, message domain.Message) (string, error) {
	cs := a.model.StartChat()

	for _, message := range a.history {
		var role string
		if message.Role == domain.UserRole {
			role = "user"
		} else {
			role = "model"
		}
		cs.History = append(cs.History, &genai.Content{
			Parts: []genai.Part{
				genai.Text(message.Content),
			},
			Role: role,
		})
	}

	a.logger.Debug("sending message", zap.String("role", "user"), zap.String("content", message.Content))

	part := genai.Text(fmt.Sprintf("[%s] %s", message.Role, message.Content))

	res, err := cs.SendMessage(ctx, part)
	if err != nil {
		var gerr *googleapi.Error
		if !errors.As(err, &gerr) {
			a.logger.Debug("failed to send message", zap.Error(err))
		} else {
			a.logger.Debug(
				"failed to send message",
				zap.Int("code", gerr.Code),
				zap.String("message", gerr.Message),
				zap.String("body", gerr.Body),
			)
		}
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	var resContent string
	for _, cand := range res.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				text, ok := part.(genai.Text)
				if !ok {
					continue
				}
				resContent += string(text)
			}
		}
	}

	response := &struct {
		ToolUse string `json:"toolUse"`
	}{}

	if err := json.Unmarshal([]byte(resContent), response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	a.logger.Debug("received response", zap.String("role", "agent"), zap.String("content", response.ToolUse))

	a.history = append(a.history, message)
	a.history = append(a.history, domain.Message{
		Role:    domain.AssistantRole,
		Content: response.ToolUse,
	})

	return response.ToolUse, nil
}

func (a *AutoAgent) GetHistory() ([]domain.Message, error) {
	return a.history, nil
}
