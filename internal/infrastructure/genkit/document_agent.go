package genkit

import (
	"context"
	"docgent-backend/internal/domain"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type DocumentAgent struct {
	model *genai.GenerativeModel
}

func NewDocumentAgent(config Config) (*DocumentAgent, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel(config.GenerativeModelName)
	return &DocumentAgent{
		model: model,
	}, nil
}

func (c *DocumentAgent) GenerateContent(ctx context.Context, input string) (domain.DocumentContent, error) {
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
組織において、仕様書、マニュアル、ガイドラインなどのドキュメントを適切に管理し、常に最新の状態に保つことが難しいという課題があります。

以下の会話内容からドキュメント化に値する要点を抜き出し、タイトル（title）と本文（content）を含むドキュメントを生成してください。
必要に応じて、適切な見出しや箇条書きを使用してください。また、会話内容を元にしているため、ドキュメントとして適切な文体で記述してください。
本文はMarkdownフォーマットで記述してください。

会話内容:
%s
`, input)

	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return domain.DocumentContent{}, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return domain.DocumentContent{}, fmt.Errorf("no response from the model")
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
						return domain.DocumentContent{}, fmt.Errorf("failed to parse model response: %w", err)
					}
					draft := domain.NewDocumentContent(result.Title, result.Content)
					return draft, nil
				}
			}
		}
	}

	return domain.DocumentContent{}, fmt.Errorf("no valid response from the model")
}
