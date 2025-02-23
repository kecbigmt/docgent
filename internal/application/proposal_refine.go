package application

import (
	"context"
	"fmt"
	"strings"

	"docgent/internal/application/port"
	"docgent/internal/application/tooluse"
	"docgent/internal/domain"
	"docgent/internal/domain/data"
	domaintooluse "docgent/internal/domain/tooluse"
)

type ProposalRefineUsecase struct {
	chatModel           domain.ChatModel
	conversationService port.ConversationService
	fileQueryService    port.FileQueryService
	fileRepository      data.FileRepository
	proposalRepository  domain.ProposalRepository
	ragCorpus           port.RAGCorpus
	remainingStepCount  int
}

type NewProposalRefineUsecaseOption func(*ProposalRefineUsecase)

func WithProposalRefineRAGCorpus(ragCorpus port.RAGCorpus) NewProposalRefineUsecaseOption {
	return func(u *ProposalRefineUsecase) {
		u.ragCorpus = ragCorpus
	}
}

func NewProposalRefineUsecase(
	chatModel domain.ChatModel,
	conversationService port.ConversationService,
	fileQueryService port.FileQueryService,
	fileRepository data.FileRepository,
	proposalRepository domain.ProposalRepository,
	options ...NewProposalRefineUsecaseOption,
) *ProposalRefineUsecase {
	workflow := &ProposalRefineUsecase{
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

func WithRemainingStepCount(remainingStepCount int) NewProposalRefineUsecaseOption {
	return func(w *ProposalRefineUsecase) {
		w.remainingStepCount = remainingStepCount
	}
}

func (w *ProposalRefineUsecase) Refine(proposalHandle domain.ProposalHandle, userFeedback string) error {
	go w.conversationService.MarkEyes()
	defer w.conversationService.RemoveEyes()

	ctx := context.Background()

	proposal, err := w.proposalRepository.GetProposal(proposalHandle)
	if err != nil {
		if err := w.conversationService.Reply("Failed to retrieve proposal"); err != nil {
			return fmt.Errorf("failed to reply error message: %w", err)
		}
		return fmt.Errorf("failed to retrieve proposal: %w", err)
	}

	tree, err := w.fileQueryService.GetTree(ctx, port.WithGetTreeRecursive())
	if err != nil {
		return fmt.Errorf("failed to get tree metadata: %w", err)
	}

	docgentRulesFile, err := getDocgentRulesFileIfExists(ctx, w.fileQueryService, tree)
	if err != nil {
		return fmt.Errorf("failed to get docgent rules file: %w", err)
	}

	var fileChanged bool

	// ハンドラーの初期化
	attemptCompleteHandler := tooluse.NewAttemptCompleteHandler(w.conversationService)
	findFileHandler := tooluse.NewFindFileHandler(ctx, w.fileQueryService)
	fileChangeHandler := tooluse.NewFileChangeHandler(ctx, w.fileRepository, &fileChanged)
	queryRAGHandler := tooluse.NewQueryRAGHandler(ctx, w.ragCorpus)

	// ツールケースの設定
	cases := domaintooluse.Cases{
		AttemptComplete: attemptCompleteHandler.Handle,
		FindFile:        findFileHandler.Handle,
		ChangeFile:      fileChangeHandler.Handle,
		QueryRAG:        queryRAGHandler.Handle,
	}

	agent := domain.NewAgent(
		w.chatModel,
		buildSystemInstructionToRefineProposal(tree, proposal, docgentRulesFile, w.ragCorpus != nil),
		cases,
	)

	task := fmt.Sprintf(`<task>
You submitted a proposal to create/update documents.
Now, you are given a user feedback.
Use query_rag to find relevant existing documents and refine the proposal based on the user feedback.
</task>
<user_feedback>
%s
</user_feedback>
`, userFeedback)

	err = agent.InitiateTaskLoop(ctx, task, w.remainingStepCount)
	if err != nil {
		if err := w.conversationService.Reply("Something went wrong while refining the proposal"); err != nil {
			return fmt.Errorf("failed to reply error message: %w", err)
		}
		return fmt.Errorf("failed to initiate task loop: %w", err)
	}

	return nil
}

func buildSystemInstructionToRefineProposal(fileTree []port.TreeMetadata, proposal domain.Proposal, docgentRulesFile *port.File, ragEnabled bool) *domain.SystemInstruction {
	var fileTreeStr strings.Builder
	for _, metadata := range fileTree {
		fileTreeStr.WriteString(fmt.Sprintf("- %s\n", metadata.Path))
	}

	var newFiles []string
	for _, diff := range proposal.Diffs {
		newFiles = append(newFiles, "- "+diff.NewName)
	}
	newFilesStr := strings.Join(newFiles, "\n")

	environments := []domain.EnvironmentContext{
		domain.NewEnvironmentContext("File tree", fileTreeStr.String()),
		domain.NewEnvironmentContext("Current proposal files", newFilesStr),
	}

	toolUses := []domaintooluse.Usage{
		domaintooluse.CreateFileUsage,
		domaintooluse.ModifyFileUsage,
		domaintooluse.DeleteFileUsage,
		domaintooluse.RenameFileUsage,
		domaintooluse.FindFileUsage,
		domaintooluse.AttemptCompleteUsage,
	}

	if docgentRulesFile != nil {
		environments = append(environments, domain.NewEnvironmentContext("User's custom instructions", fmt.Sprintf(`The following additional instructions are provided by the user.

%s`, docgentRulesFile.Content)))
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
