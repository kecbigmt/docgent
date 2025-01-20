package workflow

import (
	"context"
	"errors"
	"fmt"

	"docgent-backend/internal/domain"
	"docgent-backend/internal/domain/autoagent"
)

type ProposalRefineWorkflow struct {
	agent               autoagent.Agent
	conversationService domain.ConversationService
	fileQueryService    domain.FileQueryService
	proposalRepository  domain.ProposalRepository
	remainingStepCount  int
	nextMessage         autoagent.Message
}

type NewProposalRefineWorkflowOption func(*ProposalRefineWorkflow)

func NewProposalRefineWorkflow(
	agent autoagent.Agent,
	conversationService domain.ConversationService,
	fileQueryService domain.FileQueryService,
	proposalRepository domain.ProposalRepository,
	options ...NewProposalRefineWorkflowOption,
) *ProposalRefineWorkflow {
	workflow := &ProposalRefineWorkflow{
		agent:               agent,
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
		w.agent.SetSystemInstruction(domain.NewProposalRefineSystemPrompt(proposal, w.remainingStepCount).String())
		rawResponse, err := w.agent.SendMessage(ctx, w.nextMessage)
		if err != nil {
			go w.conversationService.Reply("Failed to generate response")
			return fmt.Errorf("failed to generate response: %w", err)
		}
		go w.conversationService.Reply(rawResponse.Message)
		currentTaskCount++

		response, err := domain.ParseResponseFromProposalRefineAgent(rawResponse)
		if err != nil {
			go w.conversationService.Reply("Failed to parse response")
			return fmt.Errorf("failed to parse response: %w", err)
		}

		go w.conversationService.Reply(response.Message)

		switch response.Type {
		case autoagent.ToolUseResponse:
			switch response.ToolType {
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
		}

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
		if errors.Is(err, domain.ErrFileNotFound) {
			return autoagent.NewMessage(autoagent.SystemRole, fmt.Sprintf("File not found: %s", params.Name)), nil
		}
		return autoagent.Message{}, fmt.Errorf("failed to find document: %w", err)
	}

	return autoagent.NewMessage(autoagent.SystemRole, fmt.Sprintf("Found document: %s\n\n%s", file.Name, file.Content)), nil
}

func (w *ProposalRefineWorkflow) handleApplyProposalDiffsTool(proposalHandle domain.ProposalHandle, rawParams interface{}) (autoagent.Message, error) {
	params, ok := rawParams.(domain.ApplyProposalDiffsToolParams)
	if !ok {
		return autoagent.Message{}, fmt.Errorf("invalid params type")
	}

	err := w.proposalRepository.ApplyProposalDiffs(proposalHandle, params.Diffs)
	if err != nil {
		return autoagent.Message{}, fmt.Errorf("failed to update proposal diffs: %w", err)
	}

	return autoagent.NewMessage(autoagent.SystemRole, "Proposal diffs applied"), nil
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
		return autoagent.Message{}, fmt.Errorf("failed to update proposal text: %w", err)
	}

	return autoagent.NewMessage(autoagent.SystemRole, "Proposal text updated"), nil
}
