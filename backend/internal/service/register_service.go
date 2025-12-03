package service

import (
	"embyhub/config"
	"embyhub/internal/dao"
	"embyhub/internal/model"
	"embyhub/internal/util"
	"embyhub/pkg/database"
	"embyhub/pkg/emby"
	"fmt"
)

type RegisterService struct {
	userDAO      *dao.UserDAO
	emailService *EmailService
	embyClient   *emby.Client
}

func NewRegisterService() *RegisterService {
	return &RegisterService{
		userDAO:      dao.NewUserDAO(),
		emailService: NewEmailService(),
		embyClient:   emby.NewClient(&config.GlobalConfig.Emby),
	}
}

// Register 用户注册（邮箱验证方式）
func (s *RegisterService) Register(req *model.RegisterRequest) (*model.RegisterResponse, error) {
	// 0. 验证密码强度
	if err := util.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	// 1. 验证邮箱验证码
	if err := s.emailService.VerifyCode(req.Email, req.Code, model.CodeTypeRegister); err != nil {
		return nil, err
	}

	// 2. 检查本地用户名是否已存在
	existingUser, _ := s.userDAO.GetByUsername(req.Username)
	if existingUser != nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 2.1 检查邮箱是否已被使用
	existingEmailUser, _ := s.userDAO.GetByEmail(req.Email)
	if existingEmailUser != nil {
		return nil, fmt.Errorf("邮箱已被使用")
	}

	// 3. 检查Emby是否已有同名用户
	var embyUserID string
	var isExistingEmbyUser bool
	existingEmbyUser, _ := s.embyClient.GetUserByName(req.Username)
	if existingEmbyUser != nil {
		// Emby已存在同名用户，关联现有用户并更新密码
		embyUserID = existingEmbyUser.ID
		isExistingEmbyUser = true
		// 更新Emby用户密码
		if err := s.embyClient.SetUserPassword(embyUserID, req.Password); err != nil {
			return nil, fmt.Errorf("更新Emby密码失败: %w", err)
		}
	} else {
		// 创建新的Emby用户
		embyUser, err := s.embyClient.CreateUser(req.Username, req.Password)
		if err != nil {
			return nil, fmt.Errorf("创建Emby用户失败: %w", err)
		}
		embyUserID = embyUser.ID
		// 设置Emby用户密码
		if err := s.embyClient.SetUserPassword(embyUserID, req.Password); err != nil {
			return nil, fmt.Errorf("设置Emby密码失败: %w", err)
		}
		// 设置Emby用户权限（受限普通用户）
		if err := s.embyClient.SetUserPolicy(embyUserID); err != nil {
			return nil, fmt.Errorf("设置Emby权限失败: %w", err)
		}
	}

	// 5. 加密密码
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 4. 创建本地用户（默认为普通用户角色ID=3）
	user := &model.User{
		Username:     req.Username,
		PasswordHash: hashedPassword,
		Email:        req.Email,
		EmbyUserID:   embyUserID,
		RoleID:       3, // 默认为普通用户角色
		Status:       1, // 启用状态
	}

	if err := s.userDAO.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 清理已使用的验证码
	database.DB.Model(&model.EmailCode{}).
		Where("email = ? AND type = ?", req.Email, model.CodeTypeRegister).
		Update("used", true)

	// 异步发送欢迎邮件
	go s.sendWelcomeEmail(req.Email, req.Username)

	// 构建返回消息
	var msg string
	if isExistingEmbyUser {
		msg = "注册成功！已关联现有Emby账号，密码已同步更新"
	} else {
		msg = "注册成功！已创建Emby账号"
	}

	return &model.RegisterResponse{
		UserID:     user.UserID,
		Username:   user.Username,
		Email:      req.Email,
		EmbyUserID: user.EmbyUserID,
		Message:    msg,
	}, nil
}

// sendWelcomeEmail 发送欢迎邮件
func (s *RegisterService) sendWelcomeEmail(emailAddr, username string) {
	emailService := NewEmailService()
	client, err := emailService.GetEmailClient()
	if err != nil {
		return
	}
	client.SendWelcomeEmail(emailAddr, username)
}
