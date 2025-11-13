package user

import (
	"time"
)

// User represents a user reference from AAA service
// This model only stores references and e-commerce specific data
type User struct {
	Id        string `json:"id" gorm:"primaryKey;type:varchar(255)"`           // Primary key - same as AAA service user ID
	Username  string `json:"username" gorm:"type:varchar(100);not null;index"` // Cached from AAA service
	Email     string `json:"email" gorm:"type:varchar(255);index"`             // Cached from AAA service
	FirstName string `json:"first_name" gorm:"type:varchar(100)"`              // Cached from AAA service
	LastName  string `json:"last_name" gorm:"type:varchar(100)"`               // Cached from AAA service
	Phone     string `json:"phone" gorm:"type:varchar(20);index"`              // Cached from AAA service

	// E-commerce specific fields
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	LastLoginAt *time.Time `json:"last_login_at"`
	Preferences string     `json:"preferences" gorm:"type:jsonb"` // User preferences for e-commerce

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (User) TableName() string {
	return "users"
}

// NewUser creates a new user reference from AAA service
func NewUser(aaaUserId, username, email, firstName, lastName, phone string) *User {
	return &User{
		Id:        aaaUserId,
		Username:  username,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		IsActive:  true,
	}
}

// GetId returns the user ID (same as AAA service user ID)
func (u *User) GetId() string {
	return u.Id
}
