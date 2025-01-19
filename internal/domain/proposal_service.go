package domain

type ProposalService struct {
	repository ProposalRepository
	agent      ProposalAgent
}

type ProposalRepository interface {
	CreateProposal(diffs Diffs, content ProposalContent) (ProposalHandle, error)
	GetProposal(handle ProposalHandle) (Proposal, error)
	NewProposalHandle(value string) ProposalHandle
	CreateComment(handle ProposalHandle, commentBody string) (Comment, error)
	UpdateProposalContent(handle ProposalHandle, content ProposalContent) error
	ApplyProposalDiffs(handle ProposalHandle, diffs Diffs) error
}

type ProposalAgent interface {
	Generate(diffs Diffs, contextDescription string) (ProposalContent, error)
}

func NewProposalService(agent ProposalAgent, repository ProposalRepository) *ProposalService {
	return &ProposalService{agent: agent, repository: repository}
}

func (s *ProposalService) Create(diffs Diffs, content ProposalContent) (ProposalHandle, error) {
	return s.repository.CreateProposal(diffs, content)
}

func (s *ProposalService) GetProposal(handle ProposalHandle) (Proposal, error) {
	return s.repository.GetProposal(handle)
}

func (s *ProposalService) GenerateContent(diffs Diffs, contextDescription string) (ProposalContent, error) {
	return s.agent.Generate(diffs, contextDescription)
}

func (s *ProposalService) CreateComment(proposalHandle ProposalHandle, commentBody string) (Comment, error) {
	return s.repository.CreateComment(proposalHandle, commentBody)
}

func (s *ProposalService) UpdateContent(handle ProposalHandle, content ProposalContent) error {
	return s.repository.UpdateProposalContent(handle, content)
}

func (s *ProposalService) ApplyDiffs(handle ProposalHandle, diffs Diffs) error {
	return s.repository.ApplyProposalDiffs(handle, diffs)
}
