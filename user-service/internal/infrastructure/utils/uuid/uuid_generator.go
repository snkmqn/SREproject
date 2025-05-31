package uuid

import "github.com/google/uuid"

type Service struct{}

func NewUUIDService() *Service {
	return &Service{}
}

func (u *Service) GenerateUUID() string {
	return uuid.New().String()
}
