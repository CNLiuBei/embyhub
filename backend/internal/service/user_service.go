package service

import (
	"errors"
	"fmt"
	"time"

	"embyhub/config"
	"embyhub/internal/dao"
	"embyhub/internal/model"
	"embyhub/internal/util"
	"embyhub/pkg/emby"
	"embyhub/pkg/redis"

	"gorm.io/gorm"
)

type UserService struct {
	userDAO    *dao.UserDAO
	roleDAO    *dao.RoleDAO
	embyClient *emby.Client
}

func NewUserService() *UserService {
	return &UserService{
		userDAO:    dao.NewUserDAO(),
		roleDAO:    dao.NewRoleDAO(),
		embyClient: emby.NewClient(&config.GlobalConfig.Emby),
	}
}

// Create 创建用户
func (s *UserService) Create(req *model.UserCreateRequest) (*model.User, error) {
	// 检查用户名是否存在
	exists, err := s.userDAO.ExistsByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否存在
	if req.Email != "" {
		exists, err := s.userDAO.ExistsByEmail(req.Email)
		if err != nil {
			return nil, fmt.Errorf("检查邮箱失败: %w", err)
		}
		if exists {
			return nil, errors.New("邮箱已存在")
		}
	}

	// 检查角色是否存在
	_, err = s.roleDAO.GetByID(req.RoleID)
	if err != nil {
		return nil, errors.New("角色不存在")
	}

	// 加密密码
	passwordHash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 同步创建Emby用户
	var embyUserID string
	if req.EmbyUserID == "" {
		// 创建Emby用户
		embyUser, err := s.embyClient.CreateUser(req.Username, req.Password)
		if err != nil {
			return nil, fmt.Errorf("创建Emby用户失败: %w", err)
		}
		// 设置密码
		if err := s.embyClient.SetUserPassword(embyUser.ID, req.Password); err != nil {
			return nil, fmt.Errorf("设置Emby密码失败: %w", err)
		}
		// 设置权限（受限普通用户）
		if err := s.embyClient.SetUserPolicy(embyUser.ID); err != nil {
			return nil, fmt.Errorf("设置Emby权限失败: %w", err)
		}
		embyUserID = embyUser.ID
	} else {
		embyUserID = req.EmbyUserID
	}

	// 创建用户
	user := &model.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Email:        req.Email,
		EmbyUserID:   embyUserID,
		RoleID:       req.RoleID,
		Status:       1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userDAO.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 查询完整用户信息（包含角色）
	return s.userDAO.GetByID(user.UserID)
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(userID int) (*model.User, error) {
	// 先从缓存获取
	cacheKey := fmt.Sprintf("emby_ums:user:info:%d", userID)
	// 如果缓存未命中，从数据库查询
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// 更新缓存（简化处理，实际应序列化JSON）
	redis.Set(cacheKey, user.Username, 1*time.Hour)

	return user, nil
}

// Update 更新用户
func (s *UserService) Update(userID int, req *model.UserUpdateRequest) (*model.User, error) {
	// 获取用户
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 更新字段
	if req.Email != "" && req.Email != user.Email {
		exists, err := s.userDAO.ExistsByEmail(req.Email)
		if err != nil {
			return nil, fmt.Errorf("检查邮箱失败: %w", err)
		}
		if exists {
			return nil, errors.New("邮箱已存在")
		}
		user.Email = req.Email
	}

	if req.EmbyUserID != "" {
		user.EmbyUserID = req.EmbyUserID
	}

	if req.RoleID > 0 {
		_, err := s.roleDAO.GetByID(req.RoleID)
		if err != nil {
			return nil, errors.New("角色不存在")
		}
		user.RoleID = req.RoleID
	}

	if req.Status != nil {
		user.Status = *req.Status
	}

	user.UpdatedAt = time.Now()

	if err := s.userDAO.Update(user); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("emby_ums:user:info:%d", userID)
	redis.Del(cacheKey)

	return s.userDAO.GetByID(userID)
}

// Delete 删除用户
func (s *UserService) Delete(userID int) error {
	// 检查用户是否存在
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 同步删除Emby用户
	if user.EmbyUserID != "" {
		if err := s.embyClient.DeleteUser(user.EmbyUserID); err != nil {
			// 记录错误但不阻止删除本地用户
			fmt.Printf("删除Emby用户失败: %v\n", err)
		}
	}

	// 删除用户
	if err := s.userDAO.Delete(userID); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("emby_ums:user:info:%d", userID)
	redis.Del(cacheKey)

	return nil
}

// List 获取用户列表
func (s *UserService) List(req *model.UserListRequest) (*model.UserListResponse, error) {
	users, total, err := s.userDAO.List(req)
	if err != nil {
		return nil, err
	}

	return &model.UserListResponse{
		Total: int(total),
		List:  users,
	}, nil
}

// ResetPassword 重置用户密码
func (s *UserService) ResetPassword(userID int, newPassword string) error {
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// 同步更新Emby密码
	if user.EmbyUserID != "" {
		if err := s.embyClient.SetUserPassword(user.EmbyUserID, newPassword); err != nil {
			return fmt.Errorf("更新Emby密码失败: %w", err)
		}
	}

	// 加密新密码
	passwordHash, err := util.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()

	return s.userDAO.Update(user)
}

// BatchUpdateStatus 批量更新用户状态
func (s *UserService) BatchUpdateStatus(userIDs []int, status int) error {
	return s.userDAO.BatchUpdateStatus(userIDs, status)
}

// SetVip 设置用户VIP
func (s *UserService) SetVip(userID int, days int) (*model.User, error) {
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 计算VIP到期时间
	now := time.Now()
	var vipExpireAt time.Time
	if user.VipExpireAt != nil && user.VipExpireAt.After(now) {
		// 如果当前VIP未过期，在原基础上增加
		vipExpireAt = user.VipExpireAt.AddDate(0, 0, days)
	} else {
		// 从现在开始计算
		vipExpireAt = now.AddDate(0, 0, days)
	}

	user.VipLevel = 1
	user.VipExpireAt = &vipExpireAt
	user.UpdatedAt = now

	if err := s.userDAO.Update(user); err != nil {
		return nil, errors.New("设置VIP失败")
	}

	return user, nil
}
