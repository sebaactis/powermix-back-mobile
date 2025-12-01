package user

import (
	"time"

	"github.com/google/uuid"
)

type UserCreate struct {
	Name            string `json:"name"  validate:"required,min=6,max=30"`
	Email           string `json:"email" validate:"required,min=5,max=30,email"`
	Password        string `json:"password" validate:"required,min=8,max=30"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,min=8,max=30,eqfield=Password"`
}

type UserUpdate struct {
	Name  *string `json:"name" validate:"min=6, max=30"`
	Email *string `json:"email" validate:"min=5, max=30,email"`
}

type UserResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	LoginAttempts int       `json:"login_attempt"`
	LockedUntil   string    `json:"locked_until"`
	StampsCounter int       `json:"stamps_counter"`
}

type UserRecoveryPassword struct {
	Email           string `json:"email" validate:"required,email,min=6,max=32"`
	Token           string `json:"token" validate:"required,min=1,max=1000"`
	Password        string `json:"password" validate:"required,min=8,max=30"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,min=8,max=30,eqfield=Password"`
}

type UserChangePassword struct {
	Password        string     `json:"currentPassword" validate:"required,min=8,max=30"`
	NewPassword     string     `json:"newPassword" validate:"required,min=8,max=30"`
	ConfirmPassword string     `json:"confirmPassword" validate:"required,min=8,max=30,eqfield=NewPassword"`
}

func ToResponse(u *User) *UserResponse {
	return &UserResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		LockedUntil:   u.Locked_until.Truncate(time.Second).String(),
		LoginAttempts: u.LoginAttempt,
		StampsCounter: u.StampsCounter,
	}
}

func ToResponseMany(users []*User) []*UserResponse {
	response := make([]*UserResponse, len(users))

	for i, u := range users {
		response[i] = ToResponse(u)
	}

	return response
}
