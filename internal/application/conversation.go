package application

import (
	"context"
	"docgent/internal/application/port"
	"docgent/internal/application/tooluse"
	"docgent/internal/domain"
	domaintooluse "docgent/internal/domain/tooluse"
	"fmt"
	"strings"
)

// ConversationUsecase は、AIエージェントとしてユーザーとの会話を実行するユースケースです。
type ConversationUsecase struct {
	chatModel           domain.ChatModel
	conversationService port.ConversationService
	fileQueryService    port.FileQueryService
	sourceRepositories  []port.SourceRepository
	ragCorpus           port.RAGCorpus
	responseFormatter   port.ResponseFormatter
	remainingStepCount  int
}

// NewConversationUsecaseOption はConversationUsecaseの初期化オプションです。
type NewConversationUsecaseOption func(*ConversationUsecase)

// WithConversationRAGCorpus はRAGコーパスを設定するオプションです。
func WithConversationRAGCorpus(ragCorpus port.RAGCorpus) NewConversationUsecaseOption {
	return func(u *ConversationUsecase) {
		u.ragCorpus = ragCorpus
	}
}

// NewConversationUsecase はConversationUsecaseを初期化します。
func NewConversationUsecase(
	chatModel domain.ChatModel,
	conversationService port.ConversationService,
	fileQueryService port.FileQueryService,
	sourceRepositories []port.SourceRepository,
	responseFormatter port.ResponseFormatter,
	options ...NewConversationUsecaseOption,
) *ConversationUsecase {
	u := &ConversationUsecase{
		chatModel:           chatModel,
		conversationService: conversationService,
		fileQueryService:    fileQueryService,
		sourceRepositories:  sourceRepositories,
		responseFormatter:   responseFormatter,
		remainingStepCount:  10, // デフォルトのステップ数
	}

	for _, option := range options {
		option(u)
	}

	return u
}

// Execute は会話ユースケースを実行します。
func (u *ConversationUsecase) Execute(ctx context.Context) error {
	go u.conversationService.MarkEyes()
	defer u.conversationService.RemoveEyes()

	// 会話履歴を取得
	chatHistory, err := u.conversationService.GetHistory()
	if err != nil {
		return fmt.Errorf("failed to get chat history: %w", err)
	}

	// SourceRepositoryManagerの初期化
	sourceRepositoryManager := port.NewSourceRepositoryManager(u.sourceRepositories)

	// ハンドラーの初期化
	attemptCompleteHandler := tooluse.NewAttemptCompleteHandler(u.conversationService, u.responseFormatter)
	findFileHandler := tooluse.NewFindFileHandler(ctx, u.fileQueryService)
	queryRAGHandler := tooluse.NewQueryRAGHandler(ctx, u.ragCorpus)
	findSourceHandler := tooluse.NewFindSourceHandler(ctx, sourceRepositoryManager)

	// ツールケースの設定
	cases := domaintooluse.Cases{
		AttemptComplete: attemptCompleteHandler.Handle,
		FindFile:        findFileHandler.Handle,
		QueryRAG:        queryRAGHandler.Handle,
		FindSource:      findSourceHandler.Handle,
	}

	// エージェントの初期化
	agent := domain.NewAgent(
		u.chatModel,
		buildSystemInstructionForConversation(u.ragCorpus != nil),
		cases,
	)

	// タスク文字列の構築
	var task strings.Builder
	task.WriteString("<task>\n")
	task.WriteString("A user has sent you a new message on chat. Respond based on the history of the most recent conversation.\n")
	if u.ragCorpus != nil {
		task.WriteString("If it is a question that requires domain-specific knowledge to answer, use tools to retrieve relevant knowledge before responding.\n")
	} else {
		task.WriteString("It may be a question that requires domain-specific knowledge to answer, but unfortunately you do not have access to that knowledge. Please respond in good faith based on your general knowledge.\n")
	}
	task.WriteString("</task>\n")
	task.WriteString(chatHistory.ToXML())

	// タスク実行ループの開始
	err = agent.InitiateTaskLoop(ctx, task.String(), u.remainingStepCount)
	if err != nil {
		if replyErr := u.conversationService.Reply("Something went wrong. Please try again later.", true); replyErr != nil {
			return fmt.Errorf("failed to reply error message: %w", replyErr)
		}
		return fmt.Errorf("failed to initiate task loop: %w", err)
	}

	return nil
}

// buildSystemInstructionForConversation は会話用のシステムプロンプトを構築します。
func buildSystemInstructionForConversation(ragEnabled bool) *domain.SystemInstruction {
	environments := []domain.EnvironmentContext{}

	if ragEnabled {
		environments = append(environments, domain.NewEnvironmentContext("Conversation Workflow", `1. UNDERSTAND the conversation context
		a. Analyze the conversation history to grasp the user's intent
		b. Accurately comprehend the current question or request
	  
	  2. RESEARCH and UTILIZE knowledge
		a. Use query_rag to search for relevant knowledge related to the question
		b. Use find_file to examine document details when necessary
		c. Use find_source to check the origin of related information and deepen understanding
	  
	  3. GENERATE appropriate response
		a. Organize collected information to create concise and accurate answers
		b. Directly address the user's question
		c. Add explanations for technical terms when necessary
		d. Use attempt_complete to respond and end the conversation`))
	}

	toolUses := []domaintooluse.Usage{
		domaintooluse.AttemptCompleteUsage,
	}

	if ragEnabled {
		toolUses = append(toolUses, domaintooluse.QueryRAGUsage)
		toolUses = append(toolUses, domaintooluse.FindFileUsage)
		toolUses = append(toolUses, domaintooluse.FindSourceUsage)
	}

	return domain.NewSystemInstruction(
		environments,
		toolUses,
	)
}
