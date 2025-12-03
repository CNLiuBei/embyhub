package dao

import (
	"embyhub/internal/model"
	"embyhub/pkg/database"
)

type PermissionDAO struct{}

func NewPermissionDAO() *PermissionDAO {
	return &PermissionDAO{}
}

// GetByID 根据ID获取权限
func (d *PermissionDAO) GetByID(permissionID int) (*model.Permission, error) {
	var permission model.Permission
	err := database.DB.Where("permission_id = ?", permissionID).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// List 获取权限列表
func (d *PermissionDAO) List() ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := database.DB.Order("permission_id ASC").Find(&permissions).Error
	return permissions, err
}

// GetByRoleID 根据角色ID获取权限列表
func (d *PermissionDAO) GetByRoleID(roleID int) ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := database.DB.Table("permissions").
		Select("permissions.*").
		Joins("INNER JOIN role_permissions ON permissions.permission_id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

// GetPermissionKeysByRoleID 根据角色ID获取权限Key列表
func (d *PermissionDAO) GetPermissionKeysByRoleID(roleID int) ([]string, error) {
	var keys []string
	err := database.DB.Table("permissions").
		Select("permissions.permission_key").
		Joins("INNER JOIN role_permissions ON permissions.permission_id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Pluck("permission_key", &keys).Error
	return keys, err
}
