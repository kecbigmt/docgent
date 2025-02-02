package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"docgent-backend/internal/domain"
	"docgent-backend/internal/domain/autoagent"
)

type ProposalRefineWorkflow struct {
	chatModel           autoagent.ChatModel
	conversationService autoagent.ConversationService
	fileQueryService    domain.FileQueryService
	proposalRepository  domain.ProposalRepository
	remainingStepCount  int
	nextMessage         autoagent.Message
}

type NewProposalRefineWorkflowOption func(*ProposalRefineWorkflow)

func NewProposalRefineWorkflow(
	chatModel autoagent.ChatModel,
	conversationService autoagent.ConversationService,
	fileQueryService domain.FileQueryService,
	proposalRepository domain.ProposalRepository,
	options ...NewProposalRefineWorkflowOption,
) *ProposalRefineWorkflow {
	workflow := &ProposalRefineWorkflow{
		chatModel:           chatModel,
		conversationService: conversationService,
		fileQueryService:    fileQueryService,
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

	var currentTaskCount int

	w.nextMessage = autoagent.NewMessage(autoagent.UserRole, userFeedback)

	for w.remainingStepCount > 0 {
		w.chatModel.SetSystemInstruction(domain.NewProposalRefineSystemPrompt(proposal, w.remainingStepCount).String())
		rawResponse, err := w.chatModel.SendMessage(ctx, w.nextMessage)
		if err != nil {
			go w.conversationService.Reply("Failed to generate response")
			return fmt.Errorf("failed to generate response: %w", err)
		}
		currentTaskCount++

		var response autoagent.Response
		if err := json.Unmarshal([]byte(rawResponse), &response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}

		proposalRefineResponse, err := domain.ParseResponseFromProposalRefineAgent(response)
		if err != nil {
			go w.conversationService.Reply("Failed to parse response")
			return fmt.Errorf("failed to parse response: %w", err)
		}

		go w.conversationService.Reply(response.Message)

		switch proposalRefineResponse.Type {
		case autoagent.ToolUseResponse:
			switch proposalRefineResponse.ToolType {
			case domain.FindFileTool:
				systemMessage, err := w.handleFindFileTool(response.ToolParams)
				if err != nil {

					go w.conversationService.Reply("Failed to handle find_file")
					return fmt.Errorf("failed to handle find_file: %s", err)
				}
				w.nextMessage = systemMessage
				continue
			case domain.UpdateProposalTextTool:
				systemMessage, err := w.handleUpdateProposalTextTool(proposalHandle, response.ToolParams)
				if err != nil {
					go w.conversationService.Reply("Failed to handle update_proposal_text")
					return fmt.Errorf("failed to handle update_proposal_text: %s", err)
				}
				w.nextMessage = systemMessage
				continue
			case domain.ApplyProposalDiffsTool:
				systemMessage, err := w.handleApplyProposalDiffsTool(proposalHandle, response.ToolParams)
				if err != nil {
					go w.conversationService.Reply("Failed to handle apply_proposal_diffs")
					return fmt.Errorf("failed to handle apply_proposal_diffs: %s", err)
				}
				w.nextMessage = systemMessage
				continue
			}
		case autoagent.CompleteResponse:
			return nil
		case autoagent.ErrorResponse:
			return nil
		}

		go w.conversationService.Reply("Max task count reached")

		return nil
	}

	return nil
}

func (w *ProposalRefineWorkflow) handleFindFileTool(rawParams interface{}) (autoagent.Message, error) {
	params, ok := rawParams.(domain.FindFileToolParams)
	if !ok {
		return autoagent.Message{}, fmt.Errorf("invalid params type")
	}

	file, err := w.fileQueryService.Find(params.Name)
	if err != nil {
		log.Printf("Failed to find document: %s", err)
		if errors.Is(err, domain.ErrFileNotFound) {
			return autoagent.NewMessage(autoagent.UserRole, fmt.Sprintf("File not found: %s", params.Name)), nil
		}
		return autoagent.NewMessage(autoagent.UserRole, fmt.Errorf("failed to find document: %w", err).Error()), nil
	}

	return autoagent.NewMessage(autoagent.UserRole, fmt.Sprintf("Found document: %s\n\n%s", file.Name, file.Content)), nil
}

func (w *ProposalRefineWorkflow) handleApplyProposalDiffsTool(proposalHandle domain.ProposalHandle, rawParams interface{}) (autoagent.Message, error) {
	params, ok := rawParams.(domain.ApplyProposalDiffsToolParams)
	if !ok {
		return autoagent.Message{}, fmt.Errorf("invalid params type")
	}

	err := w.proposalRepository.ApplyProposalDiffs(proposalHandle, params.Diffs)
	if err != nil {
		log.Printf("Failed to update proposal diffs: %s", err)
		return autoagent.NewMessage(autoagent.UserRole, fmt.Errorf("failed to update proposal diffs: %w", err).Error()), nil
	}

	return autoagent.NewMessage(autoagent.UserRole, "Proposal diffs applied"), nil
}

func (w *ProposalRefineWorkflow) handleUpdateProposalTextTool(proposalHandle domain.ProposalHandle, rawParams interface{}) (autoagent.Message, error) {
	params, ok := rawParams.(domain.UpdateProposalTextToolParams)
	if !ok {
		return autoagent.Message{}, fmt.Errorf("invalid params type")
	}

	err := w.proposalRepository.UpdateProposalContent(
		proposalHandle,
		domain.NewProposalContent(params.Title, params.Body),
	)
	if err != nil {
		log.Printf("Failed to update proposal text: %s", err)
		return autoagent.NewMessage(autoagent.UserRole, fmt.Errorf("failed to update proposal text: %w", err).Error()), nil
	}

	return autoagent.NewMessage(autoagent.UserRole, "Proposal text updated"), nil
}
