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

type ProposalRefineUsecase struct {
	chatModel           domain.ChatModel
	conversationService port.ConversationService
	fileQueryService    port.FileQueryService
	fileChangeService   port.FileChangeService
	proposalRepository  domain.ProposalRepository
	ragCorpus           port.RAGCorpus
	remainingStepCount  int
}

type NewProposalRefineUsecaseOption func(*ProposalRefineUsecase)

func NewProposalRefineUsecase(
	chatModel domain.ChatModel,
	conversationService port.ConversationService,
	fileQueryService port.FileQueryService,
	fileChangeService port.FileChangeService,
	proposalRepository domain.ProposalRepository,
	ragCorpus port.RAGCorpus,
	options ...NewProposalRefineUsecaseOption,
) *ProposalRefineUsecase {
	workflow := &ProposalRefineUsecase{
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

func WithRemainingStepCount(remainingStepCount int) NewProposalRefineUsecaseOption {
	return func(w *ProposalRefineUsecase) {
		w.remainingStepCount = remainingStepCount
	}
}

func (w *ProposalRefineUsecase) Refine(proposalHandle domain.ProposalHandle, userFeedback string) error {
	ctx := context.Background()

	proposal, err := w.proposalRepository.GetProposal(proposalHandle)
	if err != nil {
		if err := w.conversationService.Reply("Failed to retrieve proposal"); err != nil {
			return fmt.Errorf("failed to reply error message: %w", err)
		}
		return fmt.Errorf("failed to retrieve proposal: %w", err)
	}

	agent := domain.NewAgent(
		w.chatModel,
		buildSystemInstructionToRefineProposal(proposal),
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
						return "<success>File created</success>", false, nil
					},
					ModifyFile: func(c tooluse.ModifyFile) (string, bool, error) {
						err := w.fileChangeService.ModifyFile(ctx, c.Path, c.Hunks)
						if err != nil {
							return "", false, err
						}
						return "<success>File modified</success>", false, nil
					},
					RenameFile: func(c tooluse.RenameFile) (string, bool, error) {
						err := w.fileChangeService.RenameFile(ctx, c.OldPath, c.NewPath, c.Hunks)
						if err != nil {
							return "", false, err
						}
						return "<success>File renamed</success>", false, nil
					},
					DeleteFile: func(c tooluse.DeleteFile) (string, bool, error) {
						err := w.fileChangeService.DeleteFile(ctx, c.Path)
						if err != nil {
							return "", false, err
						}
						return "<success>File deleted</success>", false, nil
					},
				}
				return change.Match(cases)
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

func buildSystemInstructionToRefineProposal(proposal domain.Proposal) *domain.SystemInstruction {
	var newFiles []string
	for _, diff := range proposal.Diffs {
		newFiles = append(newFiles, "- "+diff.NewName)
	}
	newFilesStr := strings.Join(newFiles, "\n")

	systemInstruction := domain.NewSystemInstruction(
		[]domain.EnvironmentContext{
			domain.NewEnvironmentContext("Current proposal files", newFilesStr),
		},
		[]tooluse.Usage{
			tooluse.CreateFileUsage,
			tooluse.ModifyFileUsage,
			tooluse.DeleteFileUsage,
			tooluse.RenameFileUsage,
			tooluse.FindFileUsage,
			tooluse.QueryRAGUsage,
			tooluse.AttemptCompleteUsage,
		},
	)

	return systemInstruction
}
