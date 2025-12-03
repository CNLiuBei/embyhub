package service

import (
	"errors"
	"fmt"

	"embyhub/internal/dao"
	"embyhub/internal/model"

	"gorm.io/gorm"
)

type RoleService struct {
	roleDAO       *dao.RoleDAO
	permissionDAO *dao.PermissionDAO
}

func NewRoleService() *RoleService {
	return &RoleService{
		roleDAO:       dao.NewRoleDAO(),
		permissionDAO: dao.NewPermissionDAO(),
	}
}

// Create 创建角色
func (s *RoleService) Create(req *model.RoleCreateRequest) (*model.Role, error) {
	// 检查角色名是否存在
	exists, err := s.roleDAO.ExistsByName(req.RoleName)
	if err != nil {
		return nil, fmt.Errorf("检查角色名失败: %w", err)
	}
	if exists {
		return nil, errors.New("角色名已存在")
	}

	// 创建角色
	role := &model.Role{
		RoleName:    req.RoleName,
		Description: req.Description,
	}

	if err := s.roleDAO.Create(role); err != nil {
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}

	return s.roleDAO.GetByID(role.RoleID)
}

// GetByID 根据ID获取角色
func (s *RoleService) GetByID(roleID int) (*model.Role, error) {
	return s.roleDAO.GetByID(roleID)
}

// Update 更新角色
func (s *RoleService) Update(roleID int, req *model.RoleUpdateRequest) (*model.Role, error) {
	role, err := s.roleDAO.GetByID(roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		return nil, err
	}

	// 更新字段
	if req.RoleName != "" && req.RoleName != role.RoleName {
		exists, err := s.roleDAO.ExistsByName(req.RoleName)
		if err != nil {
			return nil, fmt.Errorf("检查角色名失败: %w", err)
		}
		if exists {
			return nil, errors.New("角色名已存在")
		}
		role.RoleName = req.RoleName
	}

	if req.Description != "" {
		role.Description = req.Description
	}

	if err := s.roleDAO.Update(role); err != nil {
		return nil, fmt.Errorf("更新角色失败: %w", err)
	}

	return s.roleDAO.GetByID(roleID)
}

// Delete 删除角色
func (s *RoleService) Delete(roleID int) error {
	_, err := s.roleDAO.GetByID(roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	return s.roleDAO.Delete(roleID)
}

// List 获取角色列表
func (s *RoleService) List() (*model.RoleListResponse, error) {
	roles, err := s.roleDAO.List()
	if err != nil {
		return nil, err
	}

	return &model.RoleListResponse{
		Total: len(roles),
		List:  roles,
	}, nil
}

// AssignPermissions 为角色分配权限
func (s *RoleService) AssignPermissions(roleID int, permissionIDs []int) error {
	// 检查角色是否存在
	_, err := s.roleDAO.GetByID(roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	// 验证所有权限ID是否有效
	for _, permID := range permissionIDs {
		_, err := s.permissionDAO.GetByID(permID)
		if err != nil {
			return fmt.Errorf("权限ID %d 不存在", permID)
		}
	}

	return s.roleDAO.AssignPermissions(roleID, permissionIDs)
}
