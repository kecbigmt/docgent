package workflow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"docgent-backend/internal/domain"
	"docgent-backend/internal/domain/autoagent"
	"docgent-backend/internal/domain/autoagent/tooluse"
)

type ProposalRefineWorkflow struct {
	chatModel           autoagent.ChatModel
	conversationService autoagent.ConversationService
	fileQueryService    domain.FileQueryService
	fileChangeService   domain.FileChangeService
	proposalRepository  domain.ProposalRepository
	remainingStepCount  int
}

type NewProposalRefineWorkflowOption func(*ProposalRefineWorkflow)

func NewProposalRefineWorkflow(
	chatModel autoagent.ChatModel,
	conversationService autoagent.ConversationService,
	fileQueryService domain.FileQueryService,
	fileChangeService domain.FileChangeService,
	proposalRepository domain.ProposalRepository,
	options ...NewProposalRefineWorkflowOption,
) *ProposalRefineWorkflow {
	workflow := &ProposalRefineWorkflow{
		chatModel:           chatModel,
		conversationService: conversationService,
		fileQueryService:    fileQueryService,
		fileChangeService:   fileChangeService,
		proposalRepository:  proposalRepository,
		remainingStepCount:  5,
	}

	for _, option := range options {
		option(workflow)
	}

	return workflow
}

func WithRemainingStepCount(remainingStepCount int) NewProposalRefineWorkflowOption {
	return func(w *ProposalRefineWorkflow) {
		w.remainingStepCount = remainingStepCount
	}
}

func (w *ProposalRefineWorkflow) Refine(proposalHandle domain.ProposalHandle, userFeedback string) error {
	ctx := context.Background()

	proposal, err := w.proposalRepository.GetProposal(proposalHandle)
	if err != nil {
		go w.conversationService.Reply("Failed to retrieve proposal")
		return fmt.Errorf("failed to retrieve proposal: %w", err)
	}

	agent := autoagent.NewAgent(
		w.chatModel,
		BuildSystemInstructionToRefineProposal(proposal),
		tooluse.Cases{
			AttemptComplete: func(toolUse tooluse.AttemptComplete) (string, bool, error) {
				go w.conversationService.Reply(toolUse.Message)
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
		},
	)

	task := fmt.Sprintf(`<task>
You submitted a proposal to create/update documents.
Now, you are given a user feedback.
Refine the proposal based on the user feedback.
</task>
<user_feedback>
%s
</user_feedback>
`, userFeedback)

	err = agent.InitiateTaskLoop(ctx, task, w.remainingStepCount)
	if err != nil {
		go w.conversationService.Reply("Something went wrong while refining the proposal")
		return fmt.Errorf("failed to initiate task loop: %w", err)
	}

	return nil
}

func BuildSystemInstructionToRefineProposal(proposal domain.Proposal) *autoagent.SystemInstruction {
	var newFiles []string
	for _, diff := range proposal.Diffs {
		newFiles = append(newFiles, "- "+diff.NewName)
	}
	newFilesStr := strings.Join(newFiles, "\n")

	systemInstruction := autoagent.NewSystemInstruction(
		[]autoagent.EnvironmentContext{
			autoagent.NewEnvironmentContext("Current proposal files", newFilesStr),
		},
		[]tooluse.Usage{
			tooluse.CreateFileUsage,
			tooluse.ModifyFileUsage,
			tooluse.DeleteFileUsage,
			tooluse.RenameFileUsage,
			tooluse.FindFileUsage,
			tooluse.AttemptCompleteUsage,
		},
	)

	return systemInstruction
}
