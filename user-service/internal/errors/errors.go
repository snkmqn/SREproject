package errors

import "errors"

var (
	ErrPasswordHashing = errors.New("error hashing a password")
	ErrJWTGeneration   = errors.New("error generating a JWT")
	ErrMissingName        = errors.New("product name is required")
	ErrInvalidPrice       = errors.New("product price must be greater than zero")
	ErrInvalidStock       = errors.New("product stock must be greater than zero")
	ErrMissingDescription = errors.New("product description is required")
	ErrInvalidCategoryID  = errors.New("invalid categoryID format")
	ErrInvalidToken  = errors.New("invalid or expired token")
	ErrProductNotFound         = errors.New("product not found")
	ErrInsufficientStock       = errors.New("insufficient stock")
)
