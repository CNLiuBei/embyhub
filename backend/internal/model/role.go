package model

import "time"

// Role 角色模型
type Role struct {
	RoleID      int       `gorm:"column:role_id;primaryKey;autoIncrement" json:"role_id"`
	RoleName    string    `gorm:"column:role_name;type:varchar(50);not null;uniqueIndex" json:"role_name"`
	Description string    `gorm:"column:description;type:varchar(200)" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	// 关联
	Permissions []*Permission `gorm:"many2many:role_permissions;foreignKey:RoleID;joinForeignKey:RoleID;References:PermissionID;joinReferences:PermissionID" json:"permissions,omitempty"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// RoleCreateRequest 创建角色请求
type RoleCreateRequest struct {
	RoleName    string `json:"role_name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=200"`
}

// RoleUpdateRequest 更新角色请求
type RoleUpdateRequest struct {
	RoleName    string `json:"role_name" binding:"omitempty,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=200"`
}

// RolePermissionRequest 角色权限分配请求
type RolePermissionRequest struct {
	PermissionIDs []int `json:"permission_ids" binding:"required"`
}

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	Total int     `json:"total"`
	List  []*Role `json:"list"`
}
