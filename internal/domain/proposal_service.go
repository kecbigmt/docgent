package domain

type ProposalService struct {
	repository ProposalRepository
	agent      ProposalAgent
}

type ProposalRepository interface {
	CreateProposal(proposalContent ProposalContent, increment Increment) (Proposal, error)
	NewProposalHandle(value string) ProposalHandle
	CreateComment(handle ProposalHandle, commentBody string) (Comment, error)
}

type ProposalAgent interface {
	Generate(increment Increment) (ProposalContent, error)
}

func NewProposalService(agent ProposalAgent, repository ProposalRepository) *ProposalService {
	return &ProposalService{agent: agent, repository: repository}
}

func (s *ProposalService) Create(content ProposalContent, increment Increment) (Proposal, error) {
	return s.repository.CreateProposal(content, increment)
}

func (s *ProposalService) GenerateContent(increment Increment) (ProposalContent, error) {
	return s.agent.Generate(increment)
}

func (s *ProposalService) CreateComment(proposalHandle ProposalHandle, commentBody string) (Comment, error) {
	return s.repository.CreateComment(proposalHandle, commentBody)
}
