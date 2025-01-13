package domain

type IncrementService struct {
	repository IncrementRepository
}

type IncrementRepository interface {
	CreateIncrement(increment Increment) (Increment, error)
	NewIncrementHandle(value string) IncrementHandle
	IssueIncrementHandle() (IncrementHandle, error)
	AddDocumentChangeToIncrement(increment Increment, documentChange DocumentChange) (Increment, error)
}

func NewIncrementService(repository IncrementRepository) *IncrementService {
	return &IncrementService{repository: repository}
}

func (s *IncrementService) Create(increment Increment) (Increment, error) {
	return s.repository.CreateIncrement(increment)
}

func (s *IncrementService) IssueHandle() (IncrementHandle, error) {
	return s.repository.IssueIncrementHandle()
}

func (s *IncrementService) AddDocumentChange(increment Increment, documentChange DocumentChange) (Increment, error) {
	return s.repository.AddDocumentChangeToIncrement(increment, documentChange)
}
