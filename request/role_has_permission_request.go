package request

type CreateRolePermissionInput struct {
	RoleID        int   `json:"role_id" binding:"required"`
	PermissionIDs []int `json:"permission_ids" binding:"required"`
}

type DeleteRolePermissionsInput struct {
	RoleID        int   `json:"role_id" binding:"required"`
	PermissionIDs []int `json:"permission_ids" binding:"required"`
}
