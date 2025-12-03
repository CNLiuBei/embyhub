package dao

import (
	"embyhub/internal/model"
	"embyhub/pkg/database"
)

type RoleDAO struct{}

func NewRoleDAO() *RoleDAO {
	return &RoleDAO{}
}

// Create 创建角色
func (d *RoleDAO) Create(role *model.Role) error {
	return database.DB.Create(role).Error
}

// GetByID 根据ID获取角色
func (d *RoleDAO) GetByID(roleID int) (*model.Role, error) {
	var role model.Role
	err := database.DB.Preload("Permissions").Where("role_id = ?", roleID).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetByName 根据角色名获取角色
func (d *RoleDAO) GetByName(roleName string) (*model.Role, error) {
	var role model.Role
	err := database.DB.Where("role_name = ?", roleName).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// Update 更新角色
func (d *RoleDAO) Update(role *model.Role) error {
	return database.DB.Save(role).Error
}

// Delete 删除角色
func (d *RoleDAO) Delete(roleID int) error {
	return database.DB.Delete(&model.Role{}, roleID).Error
}

// List 获取角色列表
func (d *RoleDAO) List() ([]*model.Role, error) {
	var roles []*model.Role
	err := database.DB.Preload("Permissions").Order("role_id ASC").Find(&roles).Error
	return roles, err
}

// ExistsByName 检查角色名是否存在
func (d *RoleDAO) ExistsByName(roleName string) (bool, error) {
	var count int64
	err := database.DB.Model(&model.Role{}).Where("role_name = ?", roleName).Count(&count).Error
	return count > 0, err
}

// AssignPermissions 为角色分配权限
func (d *RoleDAO) AssignPermissions(roleID int, permissionIDs []int) error {
	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除现有权限
	if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 添加新权限
	if len(permissionIDs) > 0 {
		for _, permID := range permissionIDs {
			if err := tx.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)",
				roleID, permID).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// GetPermissionIDsByRoleID 获取角色的权限ID列表
func (d *RoleDAO) GetPermissionIDsByRoleID(roleID int) ([]int, error) {
	var permissionIDs []int
	err := database.DB.Table("role_permissions").
		Where("role_id = ?", roleID).
		Pluck("permission_id", &permissionIDs).Error
	return permissionIDs, err
}
