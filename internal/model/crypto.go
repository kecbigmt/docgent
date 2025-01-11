package model

type Crypto interface {
	GenerateID() (string, error)
}
