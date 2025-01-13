package genkit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"docgent-backend/internal/domain"
)

type ProposalAgent struct {
	model *genai.GenerativeModel
}

func NewProposalAgent(config Config) (*ProposalAgent, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel(config.GenerativeModelName)
	return &ProposalAgent{
		model: model,
	}, nil
}

func (a *ProposalAgent) Generate(increment domain.Increment) (domain.ProposalContent, error) {
	ctx := context.Background()

	// Get the document content from the increment
	if len(increment.DocumentChanges) == 0 {
		return domain.ProposalContent{}, fmt.Errorf("no document changes found in increment")
	}
	documentContent := increment.DocumentChanges[0].DocumentContent

	// Set response schema
	a.model.ResponseMIMEType = "application/json"
	a.model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"title": {Type: genai.TypeString},
			"body":  {Type: genai.TypeString},
		},
		Required: []string{"title", "body"},
	}

	// Generate title and body using Gemini
	prompt := fmt.Sprintf(`
以下のドキュメント内容を元に、GitHubのPull Requestのタイトルと説明文を生成してください。

タイトル（title）は変更内容を簡潔に表現し、説明文（body）は変更の目的と影響を説明してください。
説明文はMarkdownフォーマットで記述してください。

ドキュメントタイトル: %s
ドキュメント内容:
%s
`, documentContent.Title, documentContent.Body)

	resp, err := a.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return domain.ProposalContent{}, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return domain.ProposalContent{}, fmt.Errorf("no response from the model")
	}

	var result struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					if err := json.Unmarshal([]byte(text), &result); err != nil {
						return domain.ProposalContent{}, fmt.Errorf("failed to parse model response: %w", err)
					}
					return domain.NewProposalContent(result.Title, result.Body), nil
				}
			}
		}
	}

	return domain.ProposalContent{}, fmt.Errorf("no valid response from the model")
}
