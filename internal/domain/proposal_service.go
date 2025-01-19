package domain

type ProposalService struct {
	repository ProposalRepository
	agent      ProposalAgent
}

type ProposalRepository interface {
	CreateProposal(diffs Diffs, content ProposalContent) (ProposalHandle, error)
	NewProposalHandle(value string) ProposalHandle
	CreateComment(handle ProposalHandle, commentBody string) (Comment, error)
	UpdateProposal(handle ProposalHandle, content ProposalContent) error
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

func (s *ProposalService) GenerateContent(diffs Diffs, contextDescription string) (ProposalContent, error) {
	return s.agent.Generate(diffs, contextDescription)
}

func (s *ProposalService) CreateComment(proposalHandle ProposalHandle, commentBody string) (Comment, error) {
	return s.repository.CreateComment(proposalHandle, commentBody)
}
