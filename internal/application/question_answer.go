package application

import (
	"context"
	"docgent/internal/application/port"
	"docgent/internal/domain"
	"fmt"
	"strings"
)

type QuestionAnswerUsecase struct {
	chatModel           domain.ChatModel
	conversationService port.ConversationService
	ragCorpus           port.RAGCorpus
}

type NewQuestionAnswerUsecaseOption func(*QuestionAnswerUsecase)

func WithQuestionAnswerRAGCorpus(ragCorpus port.RAGCorpus) NewQuestionAnswerUsecaseOption {
	return func(u *QuestionAnswerUsecase) {
		u.ragCorpus = ragCorpus
	}
}

func NewQuestionAnswerUsecase(chatModel domain.ChatModel, conversationService port.ConversationService, options ...NewQuestionAnswerUsecaseOption) *QuestionAnswerUsecase {
	u := &QuestionAnswerUsecase{
		chatModel:           chatModel,
		conversationService: conversationService,
	}

	for _, option := range options {
		option(u)
	}

	return u
}

func (u *QuestionAnswerUsecase) Execute(question string) error {
	go u.conversationService.MarkEyes()
	defer u.conversationService.RemoveEyes()

	ctx := context.Background()

	var systemInstruction strings.Builder
	systemInstruction.WriteString("You are a helpful assistant.")
	if u.ragCorpus != nil {
		docs, err := u.ragCorpus.Query(ctx, question, 10, 0.5)
		if err != nil {
			return err
		}
		for _, doc := range docs {
			systemInstruction.WriteString(fmt.Sprintf("<document source=%q score=%.2f>\n%s\n</document>\n", doc.Source, doc.Score, doc.Content))
		}
		systemInstruction.WriteString("\n\nAnswer the question briefly and concisely.")
	} else {
		systemInstruction.WriteString(" Unfortunately, you do not have access to any domain-specific knowledge. Answer the question based on the general knowledge.")
	}

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
