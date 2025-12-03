package dao

import (
	"embyhub/internal/model"
	"embyhub/pkg/database"
)

type UserDAO struct{}

func NewUserDAO() *UserDAO {
	return &UserDAO{}
}

// Create 创建用户
func (d *UserDAO) Create(user *model.User) error {
	return database.DB.Create(user).Error
}

// GetByID 根据ID获取用户
func (d *UserDAO) GetByID(userID int) (*model.User, error) {
	var user model.User
	err := database.DB.Preload("Role.Permissions").Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (d *UserDAO) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := database.DB.Preload("Role.Permissions").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmbyUserID 根据Emby用户ID获取用户
func (d *UserDAO) GetByEmbyUserID(embyUserID string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("emby_user_id = ?", embyUserID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (d *UserDAO) GetByEmail(email string) (*model.User, error) {
	if email == "" {
		return nil, nil
	}
	var user model.User
	err := database.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (d *UserDAO) Update(user *model.User) error {
	return database.DB.Save(user).Error
}

// Delete 删除用户
func (d *UserDAO) Delete(userID int) error {
	return database.DB.Delete(&model.User{}, userID).Error
}

// List 获取用户列表
func (d *UserDAO) List(req *model.UserListRequest) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	query := database.DB.Model(&model.User{})

	// 关键词搜索
	if req.Keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 状态筛选
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 角色筛选
	if req.RoleID > 0 {
		query = query.Where("role_id = ?", req.RoleID)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 查询列表
	err := query.Preload("Role.Permissions").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

// Count 统计用户总数
func (d *UserDAO) Count() (int64, error) {
	var count int64
	err := database.DB.Model(&model.User{}).Count(&count).Error
	return count, err
}

// CountByStatus 根据状态统计用户数
func (d *UserDAO) CountByStatus(status int) (int64, error) {
	var count int64
	err := database.DB.Model(&model.User{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// ExistsByUsername 检查用户名是否存在
func (d *UserDAO) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := database.DB.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail 检查邮箱是否存在
func (d *UserDAO) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := database.DB.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// BatchUpdateStatus 批量更新用户状态
func (d *UserDAO) BatchUpdateStatus(userIDs []int, status int) error {
	return database.DB.Model(&model.User{}).
		Where("user_id IN ?", userIDs).
		Update("status", status).Error
}

// GetActiveUsers 获取活跃用户（最近访问过的用户）
func (d *UserDAO) GetActiveUsers(limit int) ([]*model.User, error) {
	var users []*model.User

	err := database.DB.Table("users").
		Select("DISTINCT users.*").
		Joins("INNER JOIN access_records ON users.user_id = access_records.user_id").
		Where("access_records.access_time >= NOW() - INTERVAL '24 HOUR'").
		Limit(limit).
		Find(&users).Error

	return users, err
}
