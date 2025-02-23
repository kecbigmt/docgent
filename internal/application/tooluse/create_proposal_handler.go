package tooluse

import (
	"fmt"

	"docgent/internal/domain"
	"docgent/internal/domain/tooluse"
)

// CreateProposalHandler は create_proposal ツールのハンドラーの基底構造体です
type CreateProposalHandler struct {
	proposalRepository domain.ProposalRepository
	fileChanged        *bool
}

// GenerateProposalHandler は提案生成時のcreate_proposalツールのハンドラーです
type GenerateProposalHandler struct {
	*CreateProposalHandler
	proposalHandle *domain.ProposalHandle
}

func NewGenerateProposalHandler(
	proposalRepository domain.ProposalRepository,
	fileChanged *bool,
	proposalHandle *domain.ProposalHandle,
) *GenerateProposalHandler {
	return &GenerateProposalHandler{
		CreateProposalHandler: &CreateProposalHandler{
			proposalRepository: proposalRepository,
			fileChanged:        fileChanged,
		},
		proposalHandle: proposalHandle,
	}
}

func (h *GenerateProposalHandler) Handle(toolUse tooluse.CreateProposal) (string, bool, error) {
	if !*h.fileChanged {
		return "<error>No file changes. You should change files before creating a proposal.</error>", false, nil
	}
	content := domain.NewProposalContent(toolUse.Title, toolUse.Description)
	handle, err := h.proposalRepository.CreateProposal(domain.Diffs{}, content)
	if err != nil {
		return "", false, err
	}
	*h.proposalHandle = handle
	return fmt.Sprintf("<success>Proposal created: %s</success>", handle.Value), false, nil
}

// RefineProposalHandler は提案更新時のcreate_proposalツールのハンドラーです
type RefineProposalHandler struct {
	*CreateProposalHandler
	proposalHandle domain.ProposalHandle
}

func NewRefineProposalHandler(
	proposalRepository domain.ProposalRepository,
	fileChanged *bool,
	proposalHandle domain.ProposalHandle,
) *RefineProposalHandler {
	return &RefineProposalHandler{
		CreateProposalHandler: &CreateProposalHandler{
			proposalRepository: proposalRepository,
			fileChanged:        fileChanged,
		},
		proposalHandle: proposalHandle,
	}
}

func (h *RefineProposalHandler) Handle(toolUse tooluse.CreateProposal) (string, bool, error) {
	if !*h.fileChanged {
		return "<error>No file changes. You should change files before updating the proposal.</error>", false, nil
	}
	content := domain.NewProposalContent(toolUse.Title, toolUse.Description)
	if err := h.proposalRepository.UpdateProposalContent(h.proposalHandle, content); err != nil {
		return "", false, err
	}
	return "<success>Proposal updated</success>", false, nil
}
