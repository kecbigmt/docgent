package domain

type Crypto interface {
	GenerateID() (string, error)
}
