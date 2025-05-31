package validators

import "errors"

func IsValidOrderStatus(status string) bool {
	switch status {
	case "pending", "completed", "cancelled":
		return true
	default:
		return false
	}
}

func ValidateOrderStatus(status string) error {
	if !IsValidOrderStatus(status) {
		return errors.New("invalid order status")
	}
	return nil
}
