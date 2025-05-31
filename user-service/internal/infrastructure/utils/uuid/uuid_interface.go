package uuid

type Generator interface {
	GenerateUUID() string
}
