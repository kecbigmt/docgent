package genkit

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"docgent-backend/internal/domain/autoagent"
)

type AutoAgentParams struct {
	fx.In

	Logger *zap.Logger
	Config Config
}

type AutoAgent struct {
	logger  *zap.Logger
	model   *genai.GenerativeModel
	history []autoagent.Message
}

func NewAutoAgent(params AutoAgentParams) (autoagent.ChatModel, error) {
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
			"type":     {Type: genai.TypeString},
			"message":  {Type: genai.TypeString},
			"toolType": {Type: genai.TypeString},
			"toolParams": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"k": {Type: genai.TypeString},
						"v": {Type: genai.TypeString},
					},
				},
			},
		},
		Required: []string{"type", "message", "toolType", "toolParams"},
	}
	model.SetTemperature(0.1)
	model.SetTopP(0.5)
	model.SetTopK(20)

	return &AutoAgent{
		logger:  params.Logger,
		model:   model,
		history: []autoagent.Message{},
	}, nil
}

func (a *AutoAgent) SetSystemInstruction(instruction string) error {
	a.model.SystemInstruction = genai.NewUserContent(genai.Text(instruction))
	return nil
}

func (a *AutoAgent) SendMessage(ctx context.Context, message autoagent.Message) (string, error) {
	cs := a.model.StartChat()

	for _, message := range a.history {
		cs.History = append(cs.History, &genai.Content{
			Parts: []genai.Part{
				genai.Text(message.Content),
			},
			Role: message.Role.String(),
		})
	}

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

	a.logger.Debug("response", zap.String("content", resContent))

	a.history = append(a.history, message)

	return resContent, nil
}

func (a *AutoAgent) GetHistory() ([]autoagent.Message, error) {
	return a.history, nil
}
