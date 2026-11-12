package users

import "strings"

type AccountInput struct {
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
	Firstname string `json:"firstname" validate:"max=20"`
	Lastname  string `json:"lastname" validate:"max=20"`
	Phone     string `json:"phone_number" validate:"max=15"`
}

type CreateUserRequest struct {
	AccountInput
	RoleID int `json:"role_id" validate:"required,gt=0"`
}

func (input *AccountInput) Normalize() {
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
}

type AssignRoleRequest struct {
	RoleID int `json:"role_id" validate:"required,gt=0"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active disabled"`
}
