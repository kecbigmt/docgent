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
	sourceRepositories  []port.SourceRepository
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
	sourceRepositories []port.SourceRepository,
	proposalRepository domain.ProposalRepository,
	options ...NewProposalRefineUsecaseOption,
) *ProposalRefineUsecase {
	workflow := &ProposalRefineUsecase{
		chatModel:           chatModel,
		conversationService: conversationService,
		fileQueryService:    fileQueryService,
		fileRepository:      fileRepository,
		sourceRepositories:  sourceRepositories,
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

	conversationURI := w.conversationService.URI()
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

	sourceRepositoryManager := port.NewSourceRepositoryManager(w.sourceRepositories)

	var fileChanged bool

	// ハンドラーの初期化
	attemptCompleteHandler := tooluse.NewAttemptCompleteHandler(w.conversationService)
	findFileHandler := tooluse.NewFindFileHandler(ctx, w.fileQueryService)
	fileChangeHandler := tooluse.NewFileChangeHandler(ctx, w.fileRepository, &fileChanged)
	queryRAGHandler := tooluse.NewQueryRAGHandler(ctx, w.ragCorpus)
	linkSourcesHandler := tooluse.NewLinkSourcesHandler(ctx, w.fileRepository, &fileChanged)
	findSourceHandler := tooluse.NewFindSourceHandler(ctx, sourceRepositoryManager)

	// ツールケースの設定
	cases := domaintooluse.Cases{
		AttemptComplete: attemptCompleteHandler.Handle,
		FindFile:        findFileHandler.Handle,
		ChangeFile:      fileChangeHandler.Handle,
		QueryRAG:        queryRAGHandler.Handle,
		LinkSources:     linkSourcesHandler.Handle,
		FindSource:      findSourceHandler.Handle,
	}

	agent := domain.NewAgent(
		w.chatModel,
		buildSystemInstructionToRefineProposal(tree, proposal, docgentRulesFile, w.ragCorpus != nil),
		cases,
	)

	task := fmt.Sprintf(`<task>
You've submitted a proposal to create/update documents.
Now, you are given a user feedback. Refine the proposal based on the user feedback by following the proposal refinement workflow.
</task>
<user_feedback uri=%q>
%s
</user_feedback>`, conversationURI.String(), userFeedback)

	err = agent.InitiateTaskLoop(ctx, task, w.remainingStepCount)
	if err != nil {
		if err := w.conversationService.Reply("Something went wrong while refining the proposal"); err != nil {
			return fmt.Errorf("failed to reply error message: %w", err)
		}
		return fmt.Errorf("failed to initiate task loop: %w", err)
	}

	return nil
}

func buildSystemInstructionToRefineProposal(fileTree []port.TreeMetadata, proposal domain.Proposal, docgentRulesFile *data.File, ragEnabled bool) *domain.SystemInstruction {
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
		domain.NewEnvironmentContext("Approved documents file tree", fileTreeStr.String()),
		domain.NewEnvironmentContext("Current proposal files", newFilesStr),
		domain.NewEnvironmentContext("Proposal refinement workflow", `1. DISCOVER context with find_file (locate source URLs in documents)
2. UNDERSTAND original discussions with find_source (primary sources)
3. EXPAND knowledge with query_rag (secondary sources)
4. PRESERVE context when modifying documents
5. ADD new context with link_sources`),
	}

	toolUses := []domaintooluse.Usage{
		domaintooluse.CreateFileUsage,
		domaintooluse.ModifyFileUsage,
		domaintooluse.DeleteFileUsage,
		domaintooluse.RenameFileUsage,
		domaintooluse.FindFileUsage,
		domaintooluse.AttemptCompleteUsage,
		domaintooluse.LinkSourcesUsage,
		domaintooluse.FindSourceUsage,
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
