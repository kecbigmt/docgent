package genkit

import (
	"context"
	"fmt"
	"os"
	"strings"

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

func (c *Client) GenerateDocument(ctx context.Context, input string) (string, error) {
	prompt := fmt.Sprintf(`
以下の会話内容を元に、Markdownフォーマットのドキュメントを生成してください。
必要に応じて、適切な見出しや箇条書きを使用してください。

会話内容:
%s
`, input)

	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response from the model")
	}

	var parts []string
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				text, ok := part.(genai.Text)
				if !ok {
					continue
				}
				parts = append(parts, string(text))
			}
		}
	}

	responseText := strings.Join(parts, "\n")
	return responseText, nil
}
