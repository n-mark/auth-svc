package model

import (
	"github.com/google/uuid"
)

// UserStatus represents the current state of a user account.
type UserStatus string

const (
	UserStatusConfirmPending UserStatus = "CONFIRM_PENDING"
	UserStatusActive         UserStatus = "ACTIVE"
)

// User represents a user entity in the database.
type User struct {
	ID                int
	Username          string
	Email             string
	Phone             string
	PasswordHash      string
	Status            UserStatus
	ConfirmationToken *string
}

// RegisterDTO — payload for POST /register.
type RegisterDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// LoginDTO — payload for POST /login.
type LoginDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ProfileDTO — payload returned for GET /profile.
type ProfileDTO struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// UpdateProfileDTO — payload for PUT /profile.
type UpdateProfileDTO struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type UserCreatedEvent struct {
	EventId          uuid.UUID `json:"event_id"`
	EventType        string    `json:"event_type"`
	NotificationType string    `json:"notification_type"`
	Message          string    `json:"message"`
	Version          string    `json:"version"`
	Email            string    `json:"email"`
	Payload          string    `json:"payload"`
}
