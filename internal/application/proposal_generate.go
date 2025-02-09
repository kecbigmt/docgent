package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"docgent-backend/internal/application/port"
	"docgent-backend/internal/domain"
	"docgent-backend/internal/domain/tooluse"
)

type ProposalGenerateUsecase struct {
	chatModel           domain.ChatModel
	conversationService port.ConversationService
	fileQueryService    domain.FileQueryService
	fileChangeService   domain.FileChangeService
	proposalRepository  domain.ProposalRepository
	ragCorpus           port.RAGCorpus
	remainingStepCount  int
}

type NewProposalGenerateUsecaseOption func(*ProposalGenerateUsecase)

func NewProposalGenerateUsecase(
	chatModel domain.ChatModel,
	conversationService port.ConversationService,
	fileQueryService domain.FileQueryService,
	fileChangeService domain.FileChangeService,
	proposalRepository domain.ProposalRepository,
	ragCorpus port.RAGCorpus,
	options ...NewProposalGenerateUsecaseOption,
) *ProposalGenerateUsecase {
	workflow := &ProposalGenerateUsecase{
		chatModel:           chatModel,
		conversationService: conversationService,
		fileQueryService:    fileQueryService,
		fileChangeService:   fileChangeService,
		proposalRepository:  proposalRepository,
		ragCorpus:           ragCorpus,
		remainingStepCount:  5,
	}

	for _, option := range options {
		option(workflow)
	}

	return workflow
}

func (w *ProposalGenerateUsecase) Execute(ctx context.Context) (domain.ProposalHandle, error) {
	chatHistory, err := w.conversationService.GetHistory()
	if err != nil {
		return domain.ProposalHandle{}, fmt.Errorf("failed to get chat history: %w", err)
	}

	var proposalHandle domain.ProposalHandle
	var fileChanged bool

	agent := domain.NewAgent(
		w.chatModel,
		buildSystemInstructionToGenerateProposal(),
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
					if errors.Is(err, domain.ErrFileNotFound) {
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

	var chatHistoryStr strings.Builder
	for _, msg := range chatHistory {
		chatHistoryStr.WriteString(fmt.Sprintf("%s: %s\n", msg.Author, msg.Content))
	}

	task := fmt.Sprintf(`<task>
1. Analyze the chat history. Use query_rag to find relevant existing documents and find_file to check the file content.
2. Use create_file, modify_file, rename_file, delete_file to change files for the best documentation.
3. Use create_proposal to create a proposal based on the document file changes. You should contain the chat history in the proposal description as a reference.
4. Use attempt_complete to complete the task.

You should use create_proposal only after you changed files.
</task>
<chat_history>
%s
</chat_history>
`, chatHistoryStr.String())

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

func buildSystemInstructionToGenerateProposal() *domain.SystemInstruction {
	systemInstruction := domain.NewSystemInstruction(
		[]domain.EnvironmentContext{},
		[]tooluse.Usage{
			tooluse.CreateFileUsage,
			tooluse.ModifyFileUsage,
			tooluse.DeleteFileUsage,
			tooluse.RenameFileUsage,
			tooluse.FindFileUsage,
			tooluse.CreateProposalUsage,
			tooluse.QueryRAGUsage,
			tooluse.AttemptCompleteUsage,
		},
	)

	return systemInstruction
}
