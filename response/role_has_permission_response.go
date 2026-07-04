package response

type PermissionWithAssignedRole struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Assigned    bool   `json:"assigned"`
}
