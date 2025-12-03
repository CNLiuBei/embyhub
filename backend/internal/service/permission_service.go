package service

import (
	"embyhub/internal/dao"
	"embyhub/internal/model"
)

type PermissionService struct {
	permissionDAO *dao.PermissionDAO
}

func NewPermissionService() *PermissionService {
	return &PermissionService{
		permissionDAO: dao.NewPermissionDAO(),
	}
}

// List 获取权限列表
func (s *PermissionService) List() (*model.PermissionListResponse, error) {
	permissions, err := s.permissionDAO.List()
	if err != nil {
		return nil, err
	}

	return &model.PermissionListResponse{
		Total: len(permissions),
		List:  permissions,
	}, nil
}

// GetByRoleID 根据角色ID获取权限列表
func (s *PermissionService) GetByRoleID(roleID int) ([]*model.Permission, error) {
	return s.permissionDAO.GetByRoleID(roleID)
}
