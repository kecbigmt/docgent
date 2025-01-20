package genkit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"docgent-backend/internal/domain/autoagent"
)

type AutoAgent struct {
	model   *genai.GenerativeModel
	history []autoagent.Message
}

func NewAutoAgent(config Config) (autoagent.Agent, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel(config.GenerativeModelName)
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
		Required: []string{"type", "message"},
	}

	return &AutoAgent{
		model:   model,
		history: []autoagent.Message{},
	}, nil
}

func (a *AutoAgent) SetSystemInstruction(instruction string) error {
	a.model.SystemInstruction = genai.NewUserContent(genai.Text(instruction))
	return nil
}

func (a *AutoAgent) SendMessage(ctx context.Context, message autoagent.Message) (autoagent.Response, error) {
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
		return autoagent.Response{}, fmt.Errorf("failed to send message: %w", err)
	}

	var resContent string
	for _, cand := range res.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				text, ok := part.(genai.Text)
				if ok {
					continue
				}
				resContent += string(text)
			}
		}
	}

	var response autoagent.Response
	if err := json.Unmarshal([]byte(resContent), &response); err != nil {
		return autoagent.Response{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	a.history = append(a.history, message)

	return response, nil
}

func (a *AutoAgent) GetHistory() ([]autoagent.Message, error) {
	return a.history, nil
}
