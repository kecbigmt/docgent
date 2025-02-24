package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"docgent/internal/application/port"
	"docgent/internal/application/tooluse"
	"docgent/internal/domain"
	"docgent/internal/domain/data"
	domaintooluse "docgent/internal/domain/tooluse"
)

type ProposalGenerateUsecase struct {
	chatModel           domain.ChatModel
	conversationService port.ConversationService
	fileQueryService    port.FileQueryService
	fileRepository      data.FileRepository
	proposalRepository  domain.ProposalRepository
	ragCorpus           port.RAGCorpus
	remainingStepCount  int
}

type NewProposalGenerateUsecaseOption func(*ProposalGenerateUsecase)

func WithProposalGenerateRAGCorpus(ragCorpus port.RAGCorpus) NewProposalGenerateUsecaseOption {
	return func(u *ProposalGenerateUsecase) {
		u.ragCorpus = ragCorpus
	}
}

func NewProposalGenerateUsecase(
	chatModel domain.ChatModel,
	conversationService port.ConversationService,
	fileQueryService port.FileQueryService,
	fileRepository data.FileRepository,
	proposalRepository domain.ProposalRepository,
	options ...NewProposalGenerateUsecaseOption,
) *ProposalGenerateUsecase {
	workflow := &ProposalGenerateUsecase{
		chatModel:           chatModel,
		conversationService: conversationService,
		fileQueryService:    fileQueryService,
		fileRepository:      fileRepository,
		proposalRepository:  proposalRepository,
		remainingStepCount:  10,
	}

	for _, option := range options {
		option(workflow)
	}

	return workflow
}

func (w *ProposalGenerateUsecase) Execute(ctx context.Context) (domain.ProposalHandle, error) {
	go w.conversationService.MarkEyes()
	defer w.conversationService.RemoveEyes()

	conversationURI := w.conversationService.GetURI()
	chatHistory, err := w.conversationService.GetHistory()
	if err != nil {
		return domain.ProposalHandle{}, fmt.Errorf("failed to get chat history: %w", err)
	}

	tree, err := w.fileQueryService.GetTree(ctx, port.WithGetTreeRecursive())
	if err != nil {
		return domain.ProposalHandle{}, fmt.Errorf("failed to get tree metadata: %w", err)
	}

	docgentRulesFile, err := getDocgentRulesFileIfExists(ctx, w.fileQueryService, tree)
	if err != nil {
		return domain.ProposalHandle{}, fmt.Errorf("failed to get docgent rules file: %w", err)
	}

	var proposalHandle domain.ProposalHandle
	var fileChanged bool

	// ハンドラーの初期化
	attemptCompleteHandler := tooluse.NewAttemptCompleteHandler(w.conversationService)
	findFileHandler := tooluse.NewFindFileHandler(ctx, w.fileQueryService)
	fileChangeHandler := tooluse.NewFileChangeHandler(ctx, w.fileRepository, &fileChanged)
	queryRAGHandler := tooluse.NewQueryRAGHandler(ctx, w.ragCorpus)
	generateProposalHandler := tooluse.NewGenerateProposalHandler(w.proposalRepository, &fileChanged, &proposalHandle)
	linkSourcesHandler := tooluse.NewLinkSourcesHandler(ctx, w.fileRepository, &fileChanged)

	// ツールケースの設定
	cases := domaintooluse.Cases{
		AttemptComplete: attemptCompleteHandler.Handle,
		FindFile:        findFileHandler.Handle,
		ChangeFile:      fileChangeHandler.Handle,
		QueryRAG:        queryRAGHandler.Handle,
		CreateProposal:  generateProposalHandler.Handle,
		LinkSources:     linkSourcesHandler.Handle,
	}

	agent := domain.NewAgent(
		w.chatModel,
		buildSystemInstructionToGenerateProposal(conversationURI, chatHistory, tree, docgentRulesFile, w.ragCorpus != nil),
		cases,
	)

	task := `<task>
1. Analyze the chat history. Use query_rag to find relevant existing documents and find_file to check the file content.
2. Use create_file, modify_file, rename_file, delete_file to change files for the best documentation. You should set conversation uri to each file as a knowledge source using create_file or add_knowledge_sources.
3. Use create_proposal to create a proposal based on the document file changes.
4. Use attempt_complete to complete the task.

You should use create_proposal only after you changed files.
You should not use modify_file unless the file is obviously relevant to your chat history. Basically, use create_file instead.
</task>
`

	err = agent.InitiateTaskLoop(ctx, task, w.remainingStepCount)
	if err != nil {
		if err := w.conversationService.Reply("Something went wrong while generating the proposal"); err != nil {
			return domain.ProposalHandle{}, fmt.Errorf("failed to reply error message: %w", err)
		}
		return domain.ProposalHandle{}, fmt.Errorf("failed to initiate task loop: %w", err)
	}

	if proposalHandle == (domain.ProposalHandle{}) {
		return domain.ProposalHandle{}, fmt.Errorf("proposal was not created")
	}

	return proposalHandle, nil
}

func buildSystemInstructionToGenerateProposal(
	conversationURI string,
	chatHistory []port.ConversationMessage,
	fileTree []port.TreeMetadata,
	docgentRulesFile *data.File,
	ragEnabled bool,
) *domain.SystemInstruction {

	var conversationStr strings.Builder
	conversationStr.WriteString(fmt.Sprintf("<conversation uri=%q>\n", conversationURI))
	for _, msg := range chatHistory {
		conversationStr.WriteString(fmt.Sprintf("<message author=%q>\n%s\n</message>\n", msg.Author, msg.Content))
	}
	conversationStr.WriteString("</conversation>\n")

	var fileTreeStr strings.Builder
	for _, metadata := range fileTree {
		fileTreeStr.WriteString(fmt.Sprintf("- %s\n", metadata.Path))
	}

	environments := []domain.EnvironmentContext{
		domain.NewEnvironmentContext("Conversation", conversationStr.String()),
		domain.NewEnvironmentContext("File tree", fileTreeStr.String()),
	}

	if docgentRulesFile != nil {
		environments = append(environments, domain.NewEnvironmentContext("User's custom instructions", fmt.Sprintf(`The following additional instructions are provided by the user.

%s`, docgentRulesFile.Content)))
	}

	toolUses := []domaintooluse.Usage{
		domaintooluse.CreateFileUsage,
		domaintooluse.ModifyFileUsage,
		domaintooluse.DeleteFileUsage,
		domaintooluse.RenameFileUsage,
		domaintooluse.FindFileUsage,
		domaintooluse.CreateProposalUsage,
		domaintooluse.AttemptCompleteUsage,
		domaintooluse.LinkSourcesUsage,
	}

	if ragEnabled {
		toolUses = append(toolUses, domaintooluse.QueryRAGUsage)
	}

	systemInstruction := domain.NewSystemInstruction(
		environments,
		toolUses,
	)

	return systemInstruction
}

func getDocgentRulesFileIfExists(ctx context.Context, fileQueryService port.FileQueryService, fileTree []port.TreeMetadata) (*data.File, error) {
	for _, metadata := range fileTree {
		if metadata.Path == ".docgentrules" {
			file, err := fileQueryService.FindFile(ctx, ".docgentrules")
			if err != nil {
				if errors.Is(err, port.ErrFileNotFound) {
					return nil, nil
				}
				return nil, err
			}
			return &file, nil
		}
	}
	return nil, nil
}
