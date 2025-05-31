package security

type PasswordHash interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}
