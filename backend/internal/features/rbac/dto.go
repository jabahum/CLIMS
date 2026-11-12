package rbac

type SetPermissionsRequest struct {
	Permissions *[]string `json:"permissions"`
}
