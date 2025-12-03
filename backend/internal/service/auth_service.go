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
)

type AuthService struct {
	userDAO *dao.UserDAO
}

func NewAuthService() *AuthService {
	return &AuthService{
		userDAO: dao.NewUserDAO(),
	}
}

// 登录失败锁定配置
const (
	maxLoginAttempts = 5                // 最大失败次数
	lockDuration     = 15 * time.Minute // 锁定时长
)

// Login 用户登录
func (s *AuthService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	// 检查账号是否被锁定
	lockKey := fmt.Sprintf("emby_ums:login:lock:%s", req.Username)
	if locked, _ := redis.ExistsKey(lockKey); locked {
		ttl, _ := redis.TTL(lockKey)
		return nil, fmt.Errorf("账号已被锁定，请%d分钟后重试", int(ttl.Minutes())+1)
	}

	// 查询用户
	user, err := s.userDAO.GetByUsername(req.Username)
	if err != nil {
		s.recordLoginFailure(req.Username, "", "")
		return nil, errors.New("用户名或密码错误")
	}

	// 检查账号状态
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	// 验证密码
	if !util.CheckPassword(req.Password, user.PasswordHash) {
		s.recordLoginFailure(req.Username, "", "")
		return nil, errors.New("用户名或密码错误")
	}

	// 登录成功，清除失败记录
	s.clearLoginFailure(req.Username)

	// 生成Token
	token, err := util.GenerateToken(user.UserID, user.Username, user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}

	// 将Token存入Redis
	tokenKey := fmt.Sprintf("emby_ums:jwt:token:%d", user.UserID)
	if err := redis.Set(tokenKey, token, 24*time.Hour); err != nil {
		util.Warn("存储Token到Redis失败")
	}

	return &model.LoginResponse{
		Token:    token,
		UserInfo: user,
	}, nil
}

// recordLoginFailure 记录登录失败
func (s *AuthService) recordLoginFailure(username string, ip, ua string) {
	attemptsKey := fmt.Sprintf("emby_ums:login:attempts:%s", username)
	lockKey := fmt.Sprintf("emby_ums:login:lock:%s", username)

	// 增加失败次数
	attempts, _ := redis.Incr(attemptsKey)
	redis.Expire(attemptsKey, 30*time.Minute) // 30分钟内的失败次数

	// 达到最大次数则锁定
	if attempts >= maxLoginAttempts {
		redis.Set(lockKey, "1", lockDuration)
		redis.Del(attemptsKey)
		util.Warn(fmt.Sprintf("用户 %s 登录失败次数过多，已锁定%d分钟", username, int(lockDuration.Minutes())))
	}

	// 记录审计日志
	Audit(nil, username, "login_failed", "user", "", map[string]interface{}{
		"attempts": attempts,
		"reason":   "密码错误",
	}, ip, ua, "failed")
}

// clearLoginFailure 清除登录失败记录
func (s *AuthService) clearLoginFailure(username string) {
	attemptsKey := fmt.Sprintf("emby_ums:login:attempts:%s", username)
	redis.Del(attemptsKey)
}

// Logout 用户登出
func (s *AuthService) Logout(userID int) error {
	tokenKey := fmt.Sprintf("emby_ums:jwt:token:%d", userID)
	return redis.Del(tokenKey)
}

// ValidateToken 验证Token
func (s *AuthService) ValidateToken(token string) (*util.Claims, error) {
	claims, err := util.ParseToken(token)
	if err != nil {
		return nil, err
	}

	// 检查Redis中是否存在
	tokenKey := fmt.Sprintf("emby_ums:jwt:token:%d", claims.UserID)
	storedToken, err := redis.Get(tokenKey)
	if err != nil || storedToken != token {
		return nil, errors.New("Token已失效")
	}

	return claims, nil
}

// GetUserByID 根据ID获取用户（包含角色和权限）
func (s *AuthService) GetUserByID(userID int) (*model.User, error) {
	return s.userDAO.GetByID(userID)
}

// ChangePassword 用户修改自己的密码
func (s *AuthService) ChangePassword(userID int, newPassword string) error {
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 加密新密码
	hashedPassword, err := util.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败")
	}

	user.PasswordHash = hashedPassword
	if err := s.userDAO.Update(user); err != nil {
		return fmt.Errorf("更新密码失败")
	}

	// 同步更新Emby密码
	if user.EmbyUserID != "" {
		embyClient := emby.NewClient(&config.GlobalConfig.Emby)
		if err := embyClient.SetUserPassword(user.EmbyUserID, newPassword); err != nil {
			// 记录错误但不阻止操作
			fmt.Printf("同步Emby密码失败: %v\n", err)
		}
	}

	// 异步发送密码修改通知邮件
	go s.sendPasswordChangedEmail(user)

	return nil
}

// sendPasswordChangedEmail 发送密码修改通知
func (s *AuthService) sendPasswordChangedEmail(user *model.User) {
	if user.Email == "" {
		return
	}
	emailService := NewEmailService()
	client, err := emailService.GetEmailClient()
	if err != nil {
		return
	}
	client.SendPasswordChangedEmail(user.Email, user.Username, time.Now().Format("2006-01-02 15:04:05"))
}
