package users

import "clims-backend/internal/features/rbac"

type User struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Firstname    string     `json:"firstname"`
	Lastname     string     `json:"lastname"`
	Phone        string     `json:"phone_number"`
	Status       string     `json:"status"`
	Role         *rbac.Role `json:"role"`
	PasswordHash string     `json:"-"`
}

func (u *User) HasPermission(code string) bool {
	if u == nil || u.Role == nil {
		return false
	}
	for _, permission := range u.Role.Permissions {
		if permission == code {
			return true
		}
	}
	return false
}
