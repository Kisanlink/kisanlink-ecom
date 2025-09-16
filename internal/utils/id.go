package utils

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// GenerateID generates a unique ID with a prefix and sequential number.
func GenerateID(prefix string) string {
	model := base.NewBaseModel(prefix, "large")
	return model.ID
}

// GenerateUserID generates a unique user ID.
func GenerateUserID() string {
	return GenerateID("usr")
}

// GenerateProductID generates a unique product ID.
func GenerateProductID() string {
	return GenerateID("prd")
}

// GenerateOrderID generates a unique order ID.
func GenerateOrderID() string {
	return GenerateID("ord")
}
