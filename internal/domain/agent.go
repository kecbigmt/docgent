package domain

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"docgent-backend/internal/domain/tooluse"
)

type Agent struct {
	chatModel         ChatModel
	tools             tooluse.Cases
	systemInstruction *SystemInstruction
}

func NewAgent(chatModel ChatModel, systemInstruction *SystemInstruction, tools tooluse.Cases) *Agent {
	return &Agent{chatModel: chatModel, tools: tools, systemInstruction: systemInstruction}
}

func (a *Agent) InitiateTaskLoop(ctx context.Context, task string, maxStepCount int) error {
	currentStepCount := 0
	nextMessage := NewMessage(UserRole, task)
	a.chatModel.SetSystemInstruction(a.systemInstruction.String())

	for currentStepCount <= maxStepCount {
		rawResponse, err := a.chatModel.SendMessage(ctx, nextMessage)
		if err != nil {
			return fmt.Errorf("failed to generate response: %w", err)
		}
		currentStepCount++

		rawResponse = sanitizeRawResponse(rawResponse)

		toolUse, err := tooluse.Parse(rawResponse)
		if err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		message, completed, err := toolUse.Match(a.tools)
		if err != nil {
			return fmt.Errorf("failed to match tool use: %w", err)
		}
		if completed {
			return nil
		}
		nextMessage = NewMessage(UserRole, message)
	}

	return fmt.Errorf("max task count reached")
}

// 生のレスポンスからXML部分を抽出する
func sanitizeRawResponse(raw string) string {
	// 前後の空白を削除
	raw = strings.TrimSpace(raw)

	// 最初の開始タグの位置を見つける
	start := findFirstOpeningTag(raw)
	if start == -1 {
		return raw
	}

	// 最後の終了タグの位置を見つける
	end := findLastClosingTag(raw)
	if end == -1 {
		return raw
	}

	// XML部分を抽出
	return raw[start:end]
}

// 文字列の先頭から最初の開始タグの位置を返す
func findFirstOpeningTag(s string) int {
	re := regexp.MustCompile("<[^<>]+>")
	loc := re.FindStringIndex(s)
	if loc == nil {
		return -1
	}
	return loc[0]
}

// 文字列の末尾から最後の終了タグの位置を返す
func findLastClosingTag(s string) int {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	reversed := string(runes)

	re := regexp.MustCompile(">[^<>/]+/<")
	loc := re.FindStringIndex(reversed)
	if loc == nil {
		return -1
	}
	return len(s) - loc[0]
}
