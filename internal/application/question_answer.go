package application

import (
	"context"
	"docgent-backend/internal/application/port"
	"docgent-backend/internal/domain"
	"fmt"
	"strings"
)

type QuestionAnswerUsecase struct {
	chatModel           domain.ChatModel
	ragCorpus           port.RAGCorpus
	conversationService port.ConversationService
}

func NewQuestionAnswerUsecase(chatModel domain.ChatModel, ragCorpus port.RAGCorpus, conversationService port.ConversationService) *QuestionAnswerUsecase {
	return &QuestionAnswerUsecase{
		chatModel:           chatModel,
		ragCorpus:           ragCorpus,
		conversationService: conversationService,
	}
}

func (u *QuestionAnswerUsecase) Execute(question string) error {
	go u.conversationService.MarkEyes()
	defer u.conversationService.RemoveEyes()

	ctx := context.Background()

	docs, err := u.ragCorpus.Query(ctx, question, 10, 0.5)
	if err != nil {
		return err
	}

	var systemInstruction strings.Builder
	systemInstruction.WriteString("You are a helpful assistant. The following documents are selected in order of relevance to the question, but they may not necessarily be directly related.\n\n")
	for _, doc := range docs {
		systemInstruction.WriteString(fmt.Sprintf("<document source=%q score=%.2f>\n%s\n</document>\n", doc.Source, doc.Score, doc.Content))
	}
	systemInstruction.WriteString("\n\nAnswer the question briefly and concisely.")

	session := u.chatModel.StartChat(systemInstruction.String())

	response, err := session.SendMessage(ctx, question)
	if err != nil {
		return err
	}

	if err := u.conversationService.Reply(response); err != nil {
		return err
	}

	return nil
}
