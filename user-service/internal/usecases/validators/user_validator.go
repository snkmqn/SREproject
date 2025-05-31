package validators

import (
	"errors"
	"regexp"
	"unicode"
	"user-service/internal/core/models"
)

type userValidator struct{}

func NewUserValidator() UserValidator {
	return &userValidator{}
}

func (v *userValidator) Validate(user models.User) error {
	if user.Username == "" || user.Email == "" || user.Password == "" {
		return errors.New("all fields must be filled")
	}

	if !isValidEmail(user.Email) {
		return errors.New("invalid email format")
	}

	if !isValidUsername(user.Username) {
		return errors.New("username must only contain Latin letters and digits, no special characters")
	}

	if !isValidPassword(user.Password) {
		return errors.New("password must be at least 6 characters long, contain Latin letters, digits, and at least one uppercase letter")
	}

	return nil
}

func isValidEmail(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@([a-zA-Z0-9-]+\.){1,2}[a-zA-Z]{2,24}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func isValidUsername(username string) bool {
	const usernameRegex = `^[a-zA-Z0-9]+$`
	re := regexp.MustCompile(usernameRegex)
	return re.MatchString(username)
}

func isValidPassword(password string) bool {
	if len(password) < 6 {
		return false
	}

	var hasDigit, hasUpper bool
	for _, ch := range password {
		if unicode.IsDigit(ch) {
			hasDigit = true
		}
		if unicode.IsUpper(ch) {
			hasUpper = true
		}
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			return false
		}
	}

	return hasDigit && hasUpper
}
