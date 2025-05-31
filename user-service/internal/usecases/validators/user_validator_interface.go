package validators

import "user-service/internal/core/models"

type UserValidator interface {
	Validate(user models.User) error
}
