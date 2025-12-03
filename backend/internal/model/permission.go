package model

// Permission 权限模型
type Permission struct {
	PermissionID   int    `gorm:"column:permission_id;primaryKey;autoIncrement" json:"permission_id"`
	PermissionName string `gorm:"column:permission_name;type:varchar(50);not null;uniqueIndex" json:"permission_name"`
	PermissionKey  string `gorm:"column:permission_key;type:varchar(50);not null;uniqueIndex" json:"permission_key"`
	Description    string `gorm:"column:description;type:varchar(200)" json:"description"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// PermissionListResponse 权限列表响应
type PermissionListResponse struct {
	Total int           `json:"total"`
	List  []*Permission `json:"list"`
}
