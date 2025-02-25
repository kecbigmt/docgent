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
	sourceRepositories  []port.SourceRepository
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
	sourceRepositories []port.SourceRepository,
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

	conversationURI := w.conversationService.URI()

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

	sourceRepositoryManager := port.NewSourceRepositoryManager(w.sourceRepositories)

	var proposalHandle domain.ProposalHandle
	var fileChanged bool

	// ハンドラーの初期化
	attemptCompleteHandler := tooluse.NewAttemptCompleteHandler(w.conversationService)
	findFileHandler := tooluse.NewFindFileHandler(ctx, w.fileQueryService)
	fileChangeHandler := tooluse.NewFileChangeHandler(ctx, w.fileRepository, &fileChanged)
	queryRAGHandler := tooluse.NewQueryRAGHandler(ctx, w.ragCorpus)
	generateProposalHandler := tooluse.NewGenerateProposalHandler(w.proposalRepository, &fileChanged, &proposalHandle)
	linkSourcesHandler := tooluse.NewLinkSourcesHandler(ctx, w.fileRepository, &fileChanged)
	findSourceHandler := tooluse.NewFindSourceHandler(ctx, sourceRepositoryManager)

	// ツールケースの設定
	cases := domaintooluse.Cases{
		AttemptComplete: attemptCompleteHandler.Handle,
		FindFile:        findFileHandler.Handle,
		ChangeFile:      fileChangeHandler.Handle,
		QueryRAG:        queryRAGHandler.Handle,
		CreateProposal:  generateProposalHandler.Handle,
		LinkSources:     linkSourcesHandler.Handle,
		FindSource:      findSourceHandler.Handle,
	}

	agent := domain.NewAgent(
		w.chatModel,
		buildSystemInstructionToGenerateProposal(tree, docgentRulesFile, w.ragCorpus != nil),
		cases,
	)

	var task strings.Builder
	task.WriteString("<task>\n")
	task.WriteString("Create a new proposal by following the proposal generation workflow.\n")
	task.WriteString("</task>\n")
	task.WriteString(fmt.Sprintf("<conversation uri=%q>\n", conversationURI.String()))
	for _, msg := range chatHistory {
		task.WriteString(fmt.Sprintf("<message author=%q>\n%s\n</message>\n", msg.Author, msg.Content))
	}
	task.WriteString("</conversation>\n")

	err = agent.InitiateTaskLoop(ctx, task.String(), w.remainingStepCount)
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
	fileTree []port.TreeMetadata,
	docgentRulesFile *data.File,
	ragEnabled bool,
) *domain.SystemInstruction {
	var fileTreeStr strings.Builder
	for _, metadata := range fileTree {
		fileTreeStr.WriteString(fmt.Sprintf("- %s\n", metadata.Path))
	}

	environments := []domain.EnvironmentContext{
		domain.NewEnvironmentContext("Approved documents file tree", fileTreeStr.String()),
		domain.NewEnvironmentContext("Proposal generation workflow", `1. RESEARCH relevant knowledge from approved documents (secondary sources)
  a. Use query_rag to search for related existing documents
  b. Use find_file to examine full content of existing documents
  c. Determine whether to update existing documents or create new ones
2. (Optional) UNDERSTAND original discussions (primary sources) with find_source. You can find source URIs in YAML frontmatter of existing documents.
3. GENERATE document increments
  a. CREATE new documents with create_file. You should specify primary source URLs within create_file.
  b. UPDATE existing documents with modify_file, rename_file, or delete_file
  c. Add primary source URLs to the existing documents with link_sources
  d. YAML frontmatter is auto-generated, manual creation not required
4. CREATE new proposal with create_proposal. Title should be brief and descriptive. Description should be detailed and include all the changes you made and the primary source URLs. You should use create_proposal only after you changed files.
5. COMPLETE the task with attempt_complete.`),
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
		domaintooluse.FindSourceUsage,
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
