package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"docgent/internal/application/port"
	"docgent/internal/domain"
	"docgent/internal/domain/tooluse"
)

type ProposalGenerateUsecase struct {
	chatModel           domain.ChatModel
	conversationService port.ConversationService
	fileQueryService    port.FileQueryService
	fileChangeService   port.FileChangeService
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
	fileChangeService port.FileChangeService,
	proposalRepository domain.ProposalRepository,
	options ...NewProposalGenerateUsecaseOption,
) *ProposalGenerateUsecase {
	workflow := &ProposalGenerateUsecase{
		chatModel:           chatModel,
		conversationService: conversationService,
		fileQueryService:    fileQueryService,
		fileChangeService:   fileChangeService,
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

	agent := domain.NewAgent(
		w.chatModel,
		buildSystemInstructionToGenerateProposal(chatHistory, tree, docgentRulesFile, w.ragCorpus != nil),
		tooluse.Cases{
			AttemptComplete: func(toolUse tooluse.AttemptComplete) (string, bool, error) {
				if err := w.conversationService.Reply(toolUse.Message); err != nil {
					return "", false, fmt.Errorf("failed to reply: %w", err)
				}
				return "", true, nil
			},
			FindFile: func(toolUse tooluse.FindFile) (string, bool, error) {
				file, err := w.fileQueryService.FindFile(ctx, toolUse.Path)
				if err != nil {
					if errors.Is(err, port.ErrFileNotFound) {
						return fmt.Sprintf("<error>File not found: %s</error>", toolUse.Path), false, nil
					}
					return "", false, err
				}
				return fmt.Sprintf("<success>\n<content>%s</content>\n</success>", file.Content), false, nil
			},
			ChangeFile: func(toolUse tooluse.ChangeFile) (string, bool, error) {
				change := toolUse.Unwrap()
				cases := tooluse.ChangeFileCases{
					CreateFile: func(c tooluse.CreateFile) (string, bool, error) {
						err := w.fileChangeService.CreateFile(ctx, c.Path, c.Content)
						if err != nil {
							return "", false, err
						}
						fileChanged = true
						return "<success>File created</success>", false, nil
					},
					ModifyFile: func(c tooluse.ModifyFile) (string, bool, error) {
						err := w.fileChangeService.ModifyFile(ctx, c.Path, c.Hunks)
						if err != nil {
							return "", false, err
						}
						fileChanged = true
						return "<success>File modified</success>", false, nil
					},
					RenameFile: func(c tooluse.RenameFile) (string, bool, error) {
						err := w.fileChangeService.RenameFile(ctx, c.OldPath, c.NewPath, c.Hunks)
						if err != nil {
							return "", false, err
						}
						fileChanged = true
						return "<success>File renamed</success>", false, nil
					},
					DeleteFile: func(c tooluse.DeleteFile) (string, bool, error) {
						err := w.fileChangeService.DeleteFile(ctx, c.Path)
						if err != nil {
							return "", false, err
						}
						fileChanged = true
						return "<success>File deleted</success>", false, nil
					},
				}
				return change.Match(cases)
			},
			CreateProposal: func(toolUse tooluse.CreateProposal) (string, bool, error) {
				if !fileChanged {
					return "<error>No file changes. You should change files before creating a proposal.</error>", false, nil
				}
				content := domain.NewProposalContent(toolUse.Title, toolUse.Description)
				handle, err := w.proposalRepository.CreateProposal(domain.Diffs{}, content)
				if err != nil {
					return "", false, err
				}
				proposalHandle = handle
				return fmt.Sprintf("<success>Proposal created: %s</success>", handle.Value), false, nil
			},
			QueryRAG: func(toolUse tooluse.QueryRAG) (string, bool, error) {
				if w.ragCorpus == nil {
					return "<error>RAG corpus is not set.</error>", false, nil
				}
				docs, err := w.ragCorpus.Query(ctx, toolUse.Query, 10, 0.7)
				if err != nil {
					return fmt.Sprintf("<error>Failed to query RAG: %s</error>", err), false, nil
				}
				if len(docs) == 0 {
					return "<success>No relevant documents found.</success>", false, nil
				}
				var result strings.Builder
				result.WriteString("<success>\n")
				for _, doc := range docs {
					result.WriteString(fmt.Sprintf("<document source=%q score=%.2f>\n%s\n</document>\n", doc.Source, doc.Score, doc.Content))
				}
				result.WriteString("</success>")
				return result.String(), false, nil
			},
		},
	)

	task := `<task>
1. Analyze the chat history. Use query_rag to find relevant existing documents and find_file to check the file content.
2. Use create_file, modify_file, rename_file, delete_file to change files for the best documentation.
3. Use create_proposal to create a proposal based on the document file changes. You should contain the chat history in the proposal description as a reference.
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
	chatHistory []port.ConversationMessage,
	fileTree []port.TreeMetadata,
	docgentRulesFile *port.File,
	ragEnabled bool,
) *domain.SystemInstruction {
	var chatHistoryStr strings.Builder
	for _, msg := range chatHistory {
		chatHistoryStr.WriteString(fmt.Sprintf("<message author=%q>\n%s\n</message>\n", msg.Author, msg.Content))
	}

	var fileTreeStr strings.Builder
	for _, metadata := range fileTree {
		fileTreeStr.WriteString(fmt.Sprintf("- %s\n", metadata.Path))
	}

	environments := []domain.EnvironmentContext{
		domain.NewEnvironmentContext("Chat history", chatHistoryStr.String()),
		domain.NewEnvironmentContext("File tree", fileTreeStr.String()),
	}

	if docgentRulesFile != nil {
		environments = append(environments, domain.NewEnvironmentContext("User's custom instructions", fmt.Sprintf(`The following additional instructions are provided by the user.

%s`, docgentRulesFile.Content)))
	}

	toolUses := []tooluse.Usage{
		tooluse.CreateFileUsage,
		tooluse.ModifyFileUsage,
		tooluse.DeleteFileUsage,
		tooluse.RenameFileUsage,
		tooluse.FindFileUsage,
		tooluse.CreateProposalUsage,
		tooluse.AttemptCompleteUsage,
	}

	if ragEnabled {
		toolUses = append(toolUses, tooluse.QueryRAGUsage)
	}

	systemInstruction := domain.NewSystemInstruction(
		environments,
		toolUses,
	)

	return systemInstruction
}

func getDocgentRulesFileIfExists(ctx context.Context, fileQueryService port.FileQueryService, fileTree []port.TreeMetadata) (*port.File, error) {
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
