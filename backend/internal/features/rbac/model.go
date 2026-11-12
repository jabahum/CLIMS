package rbac

type Role struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}
