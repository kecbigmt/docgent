package workflow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"docgent-backend/internal/domain"
	"docgent-backend/internal/domain/tooluse"
)

type ChatMessage struct {
	Author  string
	Content string
}

type ProposalGenerateWorkflow struct {
	chatModel           domain.ChatModel
	conversationService domain.ConversationService
	fileQueryService    domain.FileQueryService
	fileChangeService   domain.FileChangeService
	proposalRepository  domain.ProposalRepository
	remainingStepCount  int
}

type NewProposalGenerateWorkflowOption func(*ProposalGenerateWorkflow)

func NewProposalGenerateWorkflow(
	chatModel domain.ChatModel,
	conversationService domain.ConversationService,
	fileQueryService domain.FileQueryService,
	fileChangeService domain.FileChangeService,
	proposalRepository domain.ProposalRepository,
	options ...NewProposalGenerateWorkflowOption,
) *ProposalGenerateWorkflow {
	workflow := &ProposalGenerateWorkflow{
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

func (w *ProposalGenerateWorkflow) Execute(
	ctx context.Context,
	chatHistory []ChatMessage,
) (domain.ProposalHandle, error) {
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
		},
	)

	var chatHistoryStr strings.Builder
	for _, msg := range chatHistory {
		chatHistoryStr.WriteString(fmt.Sprintf("%s: %s\n", msg.Author, msg.Content))
	}

	task := fmt.Sprintf(`<task>
1. Analyze the chat history. Use find_file to check the file content.
2. Use create_file, modify_file, rename_file, delete_file to change files for the best documentation.
3. Use create_proposal to create a proposal based on the document file changes. You should contain the chat history in the proposal description as a reference.
4. Use attempt_complete to complete the task.

You should use create_proposal only after you changed files.
</task>
<chat_history>
%s
</chat_history>
`, chatHistoryStr.String())

	err := agent.InitiateTaskLoop(ctx, task, w.remainingStepCount)
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
			tooluse.AttemptCompleteUsage,
		},
	)

	return systemInstruction
}
