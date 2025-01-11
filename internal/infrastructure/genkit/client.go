package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"docgent-backend/internal/model/infrastructure"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client struct {
	model *genai.GenerativeModel
}

func NewClient() (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash-001")
	return &Client{
		model: model,
	}, nil
}

func (c *Client) GenerateDocument(ctx context.Context, input string) (infrastructure.DocumentDraft, error) {
	c.model.ResponseMIMEType = "application/json"
	c.model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"title":   {Type: genai.TypeString},
			"content": {Type: genai.TypeString},
		},
		Required: []string{"title", "content"},
	}

	prompt := fmt.Sprintf(`
以下の会話内容を元に、タイトル（title）と本文（content）を含むドキュメントを生成してください。
必要に応じて、適切な見出しや箇条書きを使用してください。
本文はMarkdownフォーマットで記述してください。

会話内容:
%s
`, input)

	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return infrastructure.DocumentDraft{}, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return infrastructure.DocumentDraft{}, fmt.Errorf("no response from the model")
	}

	var result struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					if err := json.Unmarshal([]byte(text), &result); err != nil {
						return infrastructure.DocumentDraft{}, fmt.Errorf("failed to parse model response: %w", err)
					}
					draft := infrastructure.DocumentDraft{Title: result.Title, Content: result.Content}
					return draft, nil
				}
			}
		}
	}

	return infrastructure.DocumentDraft{}, fmt.Errorf("no valid response from the model")
}
