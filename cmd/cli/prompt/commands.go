package main

import (
	"encoding/json"
	"fmt"
	"os"

	"docgent-backend/cmd/cli/internal/dto"
	"docgent-backend/internal/domain"
)

type BuildCmd struct {
	ProposalFile   string `arg:"" help:"提案書データ（JSONファイル）"`
	RemainingSteps int    `flag:"" help:"残りの修正ステップ数" default:"5"`
	OutputFormat   string `flag:"" help:"出力形式 (text|json)" default:"text"`
}

func (c *BuildCmd) Run() error {
	// 提案書データの読み込み
	proposal, err := loadProposalData(c.ProposalFile)
	if err != nil {
		return fmt.Errorf("提案書データの読み込み失敗: %w", err)
	}

	// 既存のドメインロジックを活用してプロンプト生成
	prompt := domain.NewProposalRefineSystemPrompt(proposal, c.RemainingSteps)

	// フォーマットに応じて出力
	switch c.OutputFormat {
	case "json":
		return printJSON(prompt)
	default:
		fmt.Println(prompt.String())
	}
	return nil
}

func loadProposalData(filePath string) (domain.Proposal, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return domain.Proposal{}, err
	}

	var proposalDTO dto.Proposal
	if err := json.Unmarshal(data, &proposalDTO); err != nil {
		return domain.Proposal{}, err
	}

	return proposalDTO.ToDomain(), nil
}

func printJSON(prompt domain.ProposalRefineSystemPrompt) error {
	jsonData, err := json.MarshalIndent(prompt, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}
