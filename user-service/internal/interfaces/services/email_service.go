package services

type EmailService interface {
	SendWelcomeEmail(to string) error
}
