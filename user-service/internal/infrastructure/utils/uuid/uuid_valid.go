package uuid

import "github.com/google/uuid"

func IsValidUUID(id string) bool {
	parsed, err := uuid.Parse(id)
	return err == nil && parsed != uuid.Nil
}
