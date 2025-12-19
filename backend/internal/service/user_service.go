// Package service 业务逻辑层
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"feiniu-user-system/internal/config"
	"feiniu-user-system/internal/database"
	"feiniu-user-system/internal/models"
	"feiniu-user-system/pkg/auth"
	"feiniu-user-system/pkg/email"
	"feiniu-user-system/pkg/emby"
	"feiniu-user-system/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetMediaAdapterFromDB 从数据库设置获取媒体适配器
func GetMediaAdapterFromDB(db *gorm.DB) emby.MediaAdapter {
	settingService := NewSettingService(db)
	settings, err := settingService.GetEmbySettings()
	if err != nil || !settings.Enabled {
		return nil
	}

	if settings.IsEmbyMode() {
		return emby.NewEmbyAdapter(&emby.EmbyAdapterConfig{
			BaseURL: settings.BaseURL,
			APIKey:  settings.APIKey,
		})
	}
	return emby.NewFeiniuAdapter(&emby.FeiniuAdapterConfig{
		BaseURL:   settings.BaseURL,
		AdminUser: settings.AdminUser,
		AdminPass: settings.AdminPass,
	})
}

// IsEmbyEnabledFromDB 检查媒体服务是否启用
func IsEmbyEnabledFromDB(db *gorm.DB) bool {
	settingService := NewSettingService(db)
	settings, err := settingService.GetEmbySettings()
	if err != nil {
		return false
	}
	return settings.Enabled
}

// IsEmbyModeFromDB 检查是否为Emby模式
func IsEmbyModeFromDB(db *gorm.DB) bool {
	settingService := NewSettingService(db)
	settings, err := settingService.GetEmbySettings()
	if err != nil {
		return true // 默认Emby模式
	}
	return settings.IsEmbyMode()
}

// UserService 用户服务
type UserService struct {
	db         *gorm.DB
	jwtManager *auth.JWTManager
	cfg        *config.Config
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB, jwtManager *auth.JWTManager, cfg *config.Config) *UserService {
	return &UserService{db: db, jwtManager: jwtManager, cfg: cfg}
}

// getMediaAdapter 获取媒体适配器（从数据库配置）
func (s *UserService) getMediaAdapter() emby.MediaAdapter {
	return GetMediaAdapterFromDB(s.db)
}

// isEmbyEnabled 检查媒体服务是否启用
func (s *UserService) isEmbyEnabled() bool {
	return IsEmbyEnabledFromDB(s.db)
}

// LoginRequest 登录请求
type LoginRequest struct {
	Account  string `json:"account" binding:"required"` // 账号或邮箱
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	User         models.User `json:"user"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username   string `json:"username" binding:"required"`    // 账号
	Email      string `json:"email" binding:"required,email"` // 邮箱(必填)
	Code       string `json:"code" binding:"required,len=6"`  // 邮箱验证码
	Password   string `json:"password" binding:"required,min=8"`
	Nickname   string `json:"nickname"`
	InviteCode string `json:"invite_code"` // 邀请码(选填)
}

// SendRegisterCodeRequest 发送注册验证码请求
type SendRegisterCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Login 用户登录
func (s *UserService) Login(req *LoginRequest, ip string, userAgent string) (*LoginResponse, error) {
	ctx := context.Background()

	// 检查IP登录失败次数限制
	failKey := "login_fail:" + ip
	failCount, _ := database.GetCache(ctx, failKey)
	if failCount != "" {
		count := 0
		for _, c := range failCount {
			if c >= '0' && c <= '9' {
				count = count*10 + int(c-'0')
			}
		}
		if count >= 5 {
			return nil, errors.New("登录失败次数过多，请15分钟后再试")
		}
	}

	var user models.User
	// 支持账号或邮箱登录
	if err := s.db.Where("username = ? OR email = ?", req.Account, req.Account).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.incrementLoginFail(ctx, failKey)
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		s.recordLoginLog(user.ID, ip, "", 2, "密码错误")
		s.incrementLoginFail(ctx, failKey)
		return nil, errors.New("密码错误")
	}

	// 登录成功，清除失败计数
	database.DeleteCache(ctx, failKey)

	// 检查账号状态
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用，请使用卡密续费后登录")
	}

	// 检查会员状态（管理员除外）
	if user.Role < models.RoleAdmin {
		// 检查是否是会员
		if user.MemberLevel == 0 && (user.MemberExpire == nil || user.MemberExpire.Before(time.Now())) {
			return nil, errors.New("您不是会员，请先开通会员后登录")
		}
		// 检查会员是否过期
		if user.MemberExpire != nil && user.MemberExpire.Before(time.Now()) {
			// 会员已过期，禁用账户
			s.db.Model(&user).Updates(map[string]interface{}{
				"status":       2,
				"member_level": 0,
			})
			return nil, errors.New("会员已到期，账户已禁用，请使用卡密续费后登录")
		}
	}

	// 生成Token
	accessToken, err := s.jwtManager.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	// 缓存Token
	tokenKey := database.KeyUserToken + user.ID.String()
	database.SetCache(ctx, tokenKey, accessToken, time.Duration(s.cfg.JWT.AccessExpire)*time.Second)

	// 同步登录Emby并缓存token和用户ID
	if s.isEmbyEnabled() && user.Username != "" {
		if adapter := s.getMediaAdapter(); adapter != nil {
			tokenInfo, err := adapter.LoginUser(user.Username, req.Password)
			if err != nil {
				log.Printf("媒体服务登录失败: %v", err)
			} else {
				// 缓存媒体服务token
				embyTokenKey := "emby:token:" + user.ID.String()
				database.SetCache(ctx, embyTokenKey, tokenInfo.Token, 24*time.Hour)
				// 缓存Emby用户ID
				if tokenInfo.EmbyUserID != "" {
					embyUserIDKey := "emby:userid:" + user.ID.String()
					database.SetCache(ctx, embyUserIDKey, tokenInfo.EmbyUserID, 24*time.Hour)
				}
				log.Printf("媒体服务登录成功: %s", user.Username)
			}
		}
	}

	// 更新登录信息
	now := time.Now()
	s.db.Model(&user).Updates(map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": ip,
	})

	// 记录登录日志
	s.recordLoginLog(user.ID, ip, userAgent, 1, "登录成功")

	// 记录登录设备
	s.recordLoginDevice(user.ID, ip, userAgent)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.cfg.JWT.AccessExpire,
		User:         user,
	}, nil
}

// SendRegisterCode 发送注册验证码
func (s *UserService) SendRegisterCode(req *SendRegisterCodeRequest) error {
	ctx := context.Background()

	// 检查邮箱格式
	if !utils.ValidateEmail(req.Email) {
		return errors.New("邮箱格式不正确")
	}

	// 检查邮箱域名
	if !utils.ValidateEmailDomain(req.Email) {
		return errors.New("请使用常见邮箱(QQ、163、Gmail等)")
	}

	// 检查发送频率限制
	limitKey := database.KeyRegisterLimit + req.Email
	if count, _ := database.GetCache(ctx, limitKey); count != "" {
		return errors.New("发送太频繁，请稍后再试")
	}

	// 检查邮箱是否已注册
	var count int64
	s.db.Model(&models.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		return errors.New("该邮箱已注册")
	}

	// 生成6位验证码
	code := utils.GenerateCode(6)

	// 保存验证码到Redis (10分钟有效)
	codeKey := database.KeyRegisterCode + req.Email
	if err := database.SetCache(ctx, codeKey, code, 10*time.Minute); err != nil {
		return errors.New("系统繁忙，请稍后再试")
	}

	// 设置发送频率限制 (1分钟)
	database.SetCache(ctx, limitKey, "1", 1*time.Minute)

	// 发送邮件
	go func() {
		emailService := s.getEmailService()
		if emailService != nil {
			if err := emailService.SendRegisterCode(req.Email, code); err != nil {
				log.Printf("发送注册验证码失败: %v", err)
			} else {
				log.Printf("注册验证码已发送到: %s", req.Email)
			}
		}
	}()

	return nil
}

// Register 用户注册
func (s *UserService) Register(req *RegisterRequest, ip string) (*models.User, error) {
	ctx := context.Background()

	// 检查是否允许注册
	settingService := NewSettingService(s.db)
	registerSettings, _ := settingService.GetRegisterSettings()
	if registerSettings != nil && !registerSettings.Enabled {
		return nil, errors.New("系统暂不开放注册，请联系管理员")
	}

	// 验证账号格式
	if len(req.Username) < 4 || len(req.Username) > 20 {
		return nil, errors.New("账号长度必须在4-20个字符")
	}

	// 检查账号是否已注册
	var countUser int64
	s.db.Model(&models.User{}).Where("username = ?", req.Username).Count(&countUser)
	if countUser > 0 {
		return nil, errors.New("该账号已被使用")
	}

	// 验证邮箱格式和域名
	if !utils.ValidateEmail(req.Email) {
		return nil, errors.New("邮箱格式不正确")
	}
	if !utils.ValidateEmailDomain(req.Email) {
		return nil, errors.New("请使用常见邮箱(QQ、163、Gmail等)")
	}

	// 检查邮箱是否已注册
	var countEmail int64
	s.db.Model(&models.User{}).Where("email = ?", req.Email).Count(&countEmail)
	if countEmail > 0 {
		return nil, errors.New("该邮箱已注册")
	}

	// 验证邮箱验证码
	codeKey := database.KeyRegisterCode + req.Email
	savedCode, err := database.GetCache(ctx, codeKey)
	if err != nil || savedCode == "" {
		return nil, errors.New("验证码已过期，请重新获取")
	}
	if savedCode != req.Code {
		return nil, errors.New("验证码错误")
	}

	// 验证密码强度
	if !utils.ValidatePassword(req.Password, s.cfg.Security.PasswordMinLength) {
		return nil, errors.New("密码必须包含大小写字母和数字，且不少于8位")
	}

	// 在创建本地用户前，先检查媒体服务账号是否已存在
	mediaAdapter := s.getMediaAdapter()
	if s.isEmbyEnabled() && mediaAdapter != nil && req.Username != "" {
		users, err := mediaAdapter.GetUserList()
		if err != nil {
			log.Printf("无法连接媒体服务: %v", err)
		} else {
			for _, u := range users {
				if u.Username == req.Username {
					return nil, errors.New("该用户名在媒体服务中已存在，请更换用户名")
				}
			}
		}
	}

	// 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("系统错误")
	}

	// 创建用户
	nickname := req.Nickname
	if nickname == "" {
		if req.Username != "" {
			nickname = req.Username
		} else if req.Email != "" {
			nickname = req.Email[:strings.Index(req.Email, "@")]
		}
	}

	user := &models.User{
		Username:   req.Username,
		Email:      req.Email,
		Password:   hashedPassword,
		Nickname:   nickname,
		InviteCode: utils.GenerateInviteCode(), // 生成用户专属邀请码
		Status:     1,
		Role:       models.RoleUser,
		RegisterIP: ip,
	}

	// 根据设置赠送会员天数
	if registerSettings != nil && registerSettings.GiftMemberDays > 0 {
		memberExpire := time.Now().AddDate(0, 0, registerSettings.GiftMemberDays)
		user.MemberLevel = models.MemberMonth // 赠送月卡会员
		user.MemberExpire = &memberExpire
		user.Role = models.RoleMember // 升级为会员用户
		log.Printf("注册赠送会员 %d 天，到期时间: %s", registerSettings.GiftMemberDays, memberExpire.Format("2006-01-02"))
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, errors.New("注册失败，请稍后重试")
	}

	// 删除验证码
	database.DeleteCache(ctx, codeKey)

	// 处理邀请码（给邀请人奖励）
	if req.InviteCode != "" {
		inviteService := NewInviteService(s.db)
		inviteService.ProcessInviteOnRegister(user.ID, req.InviteCode)
	}

	// 同步创建媒体服务账号（注册时必须创建）
	if s.isEmbyEnabled() && mediaAdapter != nil && req.Username != "" {
		// 获取模板用户配置
		embySettings, _ := settingService.GetEmbySettings()
		templateUser := ""
		if embySettings != nil && embySettings.TemplateUser != "" {
			templateUser = embySettings.TemplateUser
		}

		var embyUserID string
		var createErr error
		if templateUser != "" {
			// 使用模板用户权限创建
			embyUserID, createErr = mediaAdapter.CreateUserWithTemplate(req.Username, req.Password, templateUser)
			if createErr != nil {
				log.Printf("使用模板用户创建媒体服务账号失败: %v, 尝试普通创建", createErr)
				// 回退到普通创建
				embyUserID, createErr = mediaAdapter.CreateUser(req.Username, req.Password)
			} else {
				log.Printf("使用模板用户[%s]创建媒体服务账号成功: %s", templateUser, req.Username)
			}
		} else {
			// 普通创建
			embyUserID, createErr = mediaAdapter.CreateUser(req.Username, req.Password)
		}

		if createErr != nil {
			log.Printf("创建媒体服务账号失败: %v", createErr)
			// 创建失败，删除本地用户并返回错误
			s.db.Delete(user)
			return nil, errors.New("创建媒体服务账号失败，请稍后重试")
		}
		
		// 保存 Emby 用户 ID 到本地数据库
		if embyUserID != "" {
			s.db.Model(user).Update("emby_user_id", embyUserID)
			user.EmbyUserID = embyUserID
		}
		
		if templateUser == "" {
			log.Printf("同步创建媒体服务账号成功: %s, embyUserID: %s", req.Username, embyUserID)
		}
	}

	// 发送欢迎邮件 (异步)
	go func() {
		emailService := s.getEmailService()
		if emailService != nil {
			if err := emailService.SendWelcome(req.Email, nickname); err != nil {
				log.Printf("发送欢迎邮件失败: %v", err)
			} else {
				log.Printf("欢迎邮件已发送到: %s", req.Email)
			}
		}
	}()

	user.Password = ""
	return user, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(userID uuid.UUID) (*models.User, error) {
	ctx := context.Background()
	cacheKey := database.KeyUserInfo + userID.String()

	// 先从缓存获取
	cached, err := database.GetCache(ctx, cacheKey)
	if err == nil {
		var user models.User
		if json.Unmarshal([]byte(cached), &user) == nil {
			return &user, nil
		}
	}

	// 从数据库获取
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 缓存用户信息
	data, _ := json.Marshal(user)
	database.SetCache(ctx, cacheKey, string(data), 10*time.Minute)

	return &user, nil
}

// UpdateUserInfo 更新用户信息
type UpdateUserRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Gender   *int8  `json:"gender"`
	Phone    string `json:"phone"`
	Bio      string `json:"bio"`
}

func (s *UserService) UpdateUserInfo(userID uuid.UUID, req *UpdateUserRequest) error {
	updates := make(map[string]interface{})
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Bio != "" {
		updates["bio"] = req.Bio
	}

	if len(updates) == 0 {
		return nil
	}

	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return errors.New("更新失败")
	}

	// 清除缓存
	ctx := context.Background()
	database.DeleteCache(ctx, database.KeyUserInfo+userID.String())

	return nil
}

// UploadAvatar 上传头像
func (s *UserService) UploadAvatar(userID uuid.UUID, file *multipart.FileHeader) (string, error) {
	// 确保上传目录存在
	uploadDir := "./uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", errors.New("创建上传目录失败")
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%d%s", userID.String(), time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, filename)

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return "", errors.New("读取文件失败")
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(savePath)
	if err != nil {
		return "", errors.New("保存文件失败")
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		return "", errors.New("保存文件失败")
	}

	// 生成访问URL
	avatarURL := "/uploads/avatars/" + filename

	// 更新用户头像
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("avatar", avatarURL).Error; err != nil {
		return "", errors.New("更新头像失败")
	}

	// 清除缓存
	ctx := context.Background()
	database.DeleteCache(ctx, database.KeyUserInfo+userID.String())

	return avatarURL, nil
}

// ChangePassword 修改密码
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (s *UserService) ChangePassword(userID uuid.UUID, req *ChangePasswordRequest) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if !utils.CheckPassword(req.OldPassword, user.Password) {
		return errors.New("原密码错误")
	}

	if !utils.ValidatePassword(req.NewPassword, s.cfg.Security.PasswordMinLength) {
		return errors.New("新密码必须包含大小写字母和数字")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("系统错误")
	}

	if err := s.db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return err
	}

	// 同步修改媒体服务密码
	if s.isEmbyEnabled() && user.Username != "" {
		go func() {
			if adapter := s.getMediaAdapter(); adapter != nil {
				if err := adapter.UpdateUserPassword(user.Username, req.NewPassword); err != nil {
					log.Printf("同步修改媒体服务密码失败: %v", err)
				} else {
					log.Printf("同步修改媒体服务密码成功: %s", user.Username)
				}
			}
		}()
	}

	// 发送密码修改通知邮件 (异步)
	if user.Email != "" {
		go func() {
			emailService := s.getEmailService()
			if emailService != nil {
				if err := emailService.SendPasswordChanged(user.Email, user.Nickname); err != nil {
					log.Printf("发送密码修改通知邮件失败: %v", err)
				}
			}
		}()
	}

	return nil
}

// Logout 退出登录
func (s *UserService) Logout(userID uuid.UUID, token string) error {
	ctx := context.Background()

	// 将Token加入黑名单
	blacklistKey := database.KeyBlacklist + token
	database.SetCache(ctx, blacklistKey, "1", time.Duration(s.cfg.JWT.AccessExpire)*time.Second)

	// 删除用户Token缓存
	database.DeleteCache(ctx, database.KeyUserToken+userID.String())

	return nil
}

// RefreshToken 刷新Token
func (s *UserService) RefreshToken(refreshToken string) (*LoginResponse, error) {
	claims, err := s.jwtManager.ParseToken(refreshToken)
	if err != nil {
		return nil, errors.New("RefreshToken无效或已过期")
	}

	var user models.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	// 生成新Token
	accessToken, _ := s.jwtManager.GenerateToken(user.ID, user.Role)
	newRefreshToken, _ := s.jwtManager.GenerateRefreshToken(user.ID, user.Role)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.cfg.JWT.AccessExpire,
	}, nil
}

// recordLoginLog 记录登录日志
// incrementLoginFail 增加登录失败计数
func (s *UserService) incrementLoginFail(ctx context.Context, key string) {
	failCount, _ := database.GetCache(ctx, key)
	count := 1
	if failCount != "" {
		for _, c := range failCount {
			if c >= '0' && c <= '9' {
				count = count*10 + int(c-'0')
			}
		}
		count++
	}
	database.SetCache(ctx, key, strconv.Itoa(count), 15*time.Minute)
}

func (s *UserService) recordLoginLog(userID uuid.UUID, ip, device string, status int8, remark string) {
	log := &models.LoginLog{
		UserID: userID,
		IP:     ip,
		Device: device,
		Status: status,
		Remark: remark,
	}
	s.db.Create(log)
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required,len=6"`
	Password string `json:"password" binding:"required,min=8"`
}

// ForgotPassword 忘记密码 - 发送验证码到邮箱
func (s *UserService) ForgotPassword(req *ForgotPasswordRequest) error {
	ctx := context.Background()

	// 检查发送频率限制 (每个邮箱1分钟只能发送1次)
	limitKey := database.KeyResetPwdLimit + req.Email
	count, _ := database.GetCache(ctx, limitKey)
	if count != "" {
		return errors.New("发送太频繁，请稍后再试")
	}

	// 检查邮箱是否存在
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("该邮箱未注册")
		}
		return err
	}

	// 检查账号状态
	if user.Status != 1 {
		return errors.New("账号已被禁用")
	}

	// 生成6位验证码
	code := utils.GenerateCode(6)

	// 保存验证码到Redis (10分钟有效)
	codeKey := database.KeyResetPwdCode + req.Email
	if err := database.SetCache(ctx, codeKey, code, 10*time.Minute); err != nil {
		return errors.New("系统繁忙，请稍后再试")
	}

	// 设置发送频率限制 (1分钟)
	database.SetCache(ctx, limitKey, "1", 1*time.Minute)

	// 发送邮件 (异步)
	go func() {
		emailService := s.getEmailService()
		if emailService != nil {
			if err := emailService.SendResetPasswordCode(req.Email, code); err != nil {
				log.Printf("发送密码重置邮件失败: %v", err)
			} else {
				log.Printf("密码重置验证码已发送到: %s", req.Email)
			}
		}
	}()

	return nil
}

// ResetPassword 重置密码 - 使用验证码重置
func (s *UserService) ResetPassword(req *ResetPasswordRequest) error {
	ctx := context.Background()

	// 获取验证码
	codeKey := database.KeyResetPwdCode + req.Email
	savedCode, err := database.GetCache(ctx, codeKey)
	if err != nil || savedCode == "" {
		return errors.New("验证码已过期，请重新获取")
	}

	// 验证验证码
	if savedCode != req.Code {
		return errors.New("验证码错误")
	}

	// 验证密码强度
	if !utils.ValidatePassword(req.Password, s.cfg.Security.PasswordMinLength) {
		return errors.New("密码必须包含大小写字母和数字，且不少于8位")
	}

	// 查找用户
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return errors.New("系统错误")
	}

	// 更新密码
	if err := s.db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return errors.New("密码重置失败")
	}

	// 删除验证码
	database.DeleteCache(ctx, codeKey)

	// 同步修改媒体服务密码
	if s.isEmbyEnabled() && user.Username != "" {
		go func() {
			if adapter := s.getMediaAdapter(); adapter != nil {
				if err := adapter.UpdateUserPassword(user.Username, req.Password); err != nil {
					log.Printf("同步修改媒体服务密码失败: %v", err)
				} else {
					log.Printf("同步修改媒体服务密码成功: %s", user.Username)
				}
			}
		}()
	}

	// 发送密码重置成功通知邮件 (异步)
	go func() {
		emailService := s.getEmailService()
		if emailService != nil {
			if err := emailService.SendPasswordChanged(req.Email, user.Nickname); err != nil {
				log.Printf("发送密码重置通知邮件失败: %v", err)
			}
		}
	}()

	return nil
}

// getEmailService 获取邮件服务 (优先使用数据库配置)
func (s *UserService) getEmailService() *email.Service {
	// 优先从数据库获取邮件配置
	settingService := NewSettingService(s.db)
	dbSettings, err := settingService.GetEmailSettings()
	if err == nil && dbSettings.Enabled && dbSettings.Host != "" {
		return email.NewServiceWithConfig(&email.Config{
			Host:     dbSettings.Host,
			Port:     dbSettings.Port,
			Username: dbSettings.Username,
			Password: dbSettings.Password,
			From:     dbSettings.From,
			FromName: dbSettings.FromName,
		})
	}

	// 回退到配置文件
	if !s.cfg.Email.Enabled {
		return nil
	}
	return email.NewService(&s.cfg.Email)
}

// recordLoginDevice 记录登录设备
func (s *UserService) recordLoginDevice(userID uuid.UUID, ip string, userAgent string) {
	// 解析User-Agent获取设备信息
	deviceName, deviceType := parseUserAgent(userAgent)
	deviceID := generateDeviceID(userID.String(), userAgent)

	now := time.Now()
	var device models.UserDevice
	err := s.db.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&device).Error
	if err != nil {
		// 新设备，创建记录
		device = models.UserDevice{
			UserID:     userID,
			DeviceID:   deviceID,
			DeviceName: deviceName,
			DeviceType: deviceType,
			LastIP:     ip,
			LastUsedAt: now,
		}
		s.db.Create(&device)
	} else {
		// 更新现有设备信息
		s.db.Model(&device).Updates(map[string]interface{}{
			"last_ip":      ip,
			"last_used_at": now,
		})
	}
}

// parseUserAgent 解析User-Agent
func parseUserAgent(ua string) (name string, deviceType string) {
	ua = strings.ToLower(ua)

	// 判断设备类型
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		deviceType = "mobile"
	} else if strings.Contains(ua, "ipad") || strings.Contains(ua, "tablet") {
		deviceType = "tablet"
	} else {
		deviceType = "desktop"
	}

	// 判断设备名称
	if strings.Contains(ua, "windows") {
		name = "Windows PC"
	} else if strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os") {
		name = "Mac"
	} else if strings.Contains(ua, "iphone") {
		name = "iPhone"
	} else if strings.Contains(ua, "ipad") {
		name = "iPad"
	} else if strings.Contains(ua, "android") {
		name = "Android"
	} else if strings.Contains(ua, "linux") {
		name = "Linux"
	} else {
		name = "未知设备"
	}

	// 添加浏览器信息
	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edge") {
		name += " Chrome"
	} else if strings.Contains(ua, "firefox") {
		name += " Firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		name += " Safari"
	} else if strings.Contains(ua, "edge") {
		name += " Edge"
	}

	return
}

// generateDeviceID 生成设备ID
func generateDeviceID(userID string, userAgent string) string {
	h := sha256.New()
	h.Write([]byte(userID + userAgent))
	return hex.EncodeToString(h.Sum(nil))[:32]
}

// SendChangeEmailCodeRequest 发送修改邮箱验证码请求
type SendChangeEmailCodeRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

// ChangeEmailRequest 修改邮箱请求
type ChangeEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
	Code     string `json:"code" binding:"required,len=6"`
}

// SendChangeEmailCode 发送修改邮箱验证码
func (s *UserService) SendChangeEmailCode(userID uuid.UUID, req *SendChangeEmailCodeRequest) error {
	ctx := context.Background()

	// 检查邮箱格式
	if !utils.ValidateEmail(req.NewEmail) {
		return errors.New("邮箱格式不正确")
	}

	// 检查邮箱域名
	if !utils.ValidateEmailDomain(req.NewEmail) {
		return errors.New("请使用常见邮箱(QQ、163、Gmail等)")
	}

	// 检查发送频率限制
	limitKey := "change_email_limit:" + userID.String()
	if count, _ := database.GetCache(ctx, limitKey); count != "" {
		return errors.New("发送太频繁，请稍后再试")
	}

	// 检查新邮箱是否已被其他用户使用
	var count int64
	s.db.Model(&models.User{}).Where("email = ? AND id != ?", req.NewEmail, userID).Count(&count)
	if count > 0 {
		return errors.New("该邮箱已被其他用户使用")
	}

	// 获取当前用户信息
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 检查是否与当前邮箱相同
	if user.Email == req.NewEmail {
		return errors.New("新邮箱与当前邮箱相同")
	}

	// 生成6位验证码
	code := utils.GenerateCode(6)

	// 保存验证码到Redis (10分钟有效)
	codeKey := "change_email_code:" + userID.String()
	emailKey := "change_email_target:" + userID.String()
	if err := database.SetCache(ctx, codeKey, code, 10*time.Minute); err != nil {
		return errors.New("系统繁忙，请稍后再试")
	}
	// 保存目标邮箱
	database.SetCache(ctx, emailKey, req.NewEmail, 10*time.Minute)

	// 设置发送频率限制 (1分钟)
	database.SetCache(ctx, limitKey, "1", 1*time.Minute)

	// 发送邮件到新邮箱
	go func() {
		emailService := s.getEmailService()
		if emailService != nil {
			if err := emailService.SendChangeEmailCode(req.NewEmail, code); err != nil {
				log.Printf("发送修改邮箱验证码失败: %v", err)
			} else {
				log.Printf("修改邮箱验证码已发送到: %s", req.NewEmail)
			}
		}
	}()

	return nil
}

// ChangeEmail 修改邮箱
func (s *UserService) ChangeEmail(userID uuid.UUID, req *ChangeEmailRequest) error {
	ctx := context.Background()

	// 检查邮箱格式
	if !utils.ValidateEmail(req.NewEmail) {
		return errors.New("邮箱格式不正确")
	}

	// 获取验证码
	codeKey := "change_email_code:" + userID.String()
	emailKey := "change_email_target:" + userID.String()
	savedCode, err := database.GetCache(ctx, codeKey)
	if err != nil || savedCode == "" {
		return errors.New("验证码已过期，请重新获取")
	}

	// 验证验证码
	if savedCode != req.Code {
		return errors.New("验证码错误")
	}

	// 验证目标邮箱是否一致
	targetEmail, _ := database.GetCache(ctx, emailKey)
	if targetEmail != req.NewEmail {
		return errors.New("邮箱不匹配，请重新获取验证码")
	}

	// 检查新邮箱是否已被其他用户使用
	var count int64
	s.db.Model(&models.User{}).Where("email = ? AND id != ?", req.NewEmail, userID).Count(&count)
	if count > 0 {
		return errors.New("该邮箱已被其他用户使用")
	}

	// 获取当前用户信息
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	oldEmail := user.Email

	// 更新邮箱
	if err := s.db.Model(&user).Update("email", req.NewEmail).Error; err != nil {
		return errors.New("修改邮箱失败")
	}

	// 删除验证码
	database.DeleteCache(ctx, codeKey)
	database.DeleteCache(ctx, emailKey)

	// 清除用户缓存
	database.DeleteCache(ctx, database.KeyUserInfo+userID.String())

	// 发送通知邮件到旧邮箱 (异步)
	if oldEmail != "" {
		go func() {
			emailService := s.getEmailService()
			if emailService != nil {
				if err := emailService.SendEmailChanged(oldEmail, user.Nickname, req.NewEmail); err != nil {
					log.Printf("发送邮箱变更通知失败: %v", err)
				}
			}
		}()
	}

	return nil
}
