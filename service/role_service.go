package service

import (
	"context"
	"errors"
	"fmt"
	"mysql/config"
	"mysql/model"
	"mysql/request"
	"mysql/response"

	"gorm.io/gorm"
)

type RoleService interface {
	CreateRoleHasPermission(ctx context.Context, input request.CreateRolePermissionInput) error
	DeleteRoleHasPermission(ctx context.Context, input request.DeleteRolePermissionsInput) error
	GetRolePermission(ctx context.Context, id int) ([]response.PermissionWithAssignedRole, error)
}

type roleservice struct {
	db *gorm.DB
}

func NewRoleService() RoleService {
	return &roleservice{
		db: config.DB,
	}
}

func (s *roleservice) CreateRoleHasPermission(ctx context.Context, input request.CreateRolePermissionInput) error {
	if len(input.PermissionIDs) == 0 {
		return errors.New("permission_ids cannot be empty")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var rolePermissions []model.RoleHasPermission
	for _, permissionID := range input.PermissionIDs {
		rolePermissions = append(rolePermissions, model.RoleHasPermission{
			RoleID:       uint(input.RoleID),
			PermissionID: uint(permissionID),
		})
	}

	if err := tx.Create(&rolePermissions).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *roleservice) DeleteRoleHasPermission(ctx context.Context, input request.DeleteRolePermissionsInput) error {
	if len(input.PermissionIDs) == 0 {
		return errors.New("permission_ids cannot be empty")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.
		Where("role_id = ? AND permission_id IN ?", input.RoleID, input.PermissionIDs).
		Delete(&model.RoleHasPermission{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *roleservice) GetRolePermission(ctx context.Context, id int) ([]response.PermissionWithAssignedRole, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid role id: %d", id)
	}

	var permissions []response.PermissionWithAssignedRole
	err := s.db.WithContext(ctx).Table("permissions").
		Select(`
            permissions.id AS id,
            permissions.name AS name,
            permissions.display_name AS display_name,
            CASE 
                WHEN role_has_permissions.permission_id IS NULL THEN false 
                ELSE true 
            END AS assigned
        `).
		Joins(`
            LEFT JOIN role_has_permissions 
            ON permissions.id = role_has_permissions.permission_id 
            AND role_has_permissions.role_id = ?
        `, id).
		Order("permissions.id ASC").
		Scan(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
