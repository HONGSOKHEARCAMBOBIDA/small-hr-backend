package model

type RoleHasPermission struct {
	RoleID       uint `json:"role_id"`
	PermissionID uint `json:"permission_id"`
}

func (RoleHasPermission) TableName() string {
	return "role_permission"
}
