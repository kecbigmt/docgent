package infrastructure

type Crypto interface {
	GenerateID() (string, error)
}
