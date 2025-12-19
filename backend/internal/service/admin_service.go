// Package service 管理员服务
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"feiniu-user-system/internal/config"
	"feiniu-user-system/internal/database"
	"feiniu-user-system/internal/models"
	"feiniu-user-system/pkg/emby"
	"feiniu-user-system/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminService 管理员服务
type AdminService struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewAdminService 创建管理员服务
func NewAdminService(db *gorm.DB, cfg *config.Config) *AdminService {
	return &AdminService{db: db, cfg: cfg}
}

// getMediaAdapter 获取媒体适配器（从数据库配置）
func (s *AdminService) getMediaAdapter() emby.MediaAdapter {
	return GetMediaAdapterFromDB(s.db)
}

// isEmbyEnabled 检查媒体服务是否启用
func (s *AdminService) isEmbyEnabled() bool {
	return IsEmbyEnabledFromDB(s.db)
}

// UserListRequest 用户列表请求
type UserListRequest struct {
	Page        int    `form:"page" binding:"required,min=1"`
	PageSize    int    `form:"page_size" binding:"required,min=1,max=100"`
	Keyword     string `form:"keyword"`  // 账号/邮箱/昵称搜索（通用）
	Username    string `form:"username"` // 账号搜索（精确）
	Email       string `form:"email"`    // 邮箱搜索（精确）
	Nickname    string `form:"nickname"` // 昵称搜索（精确）
	Role        *int8  `form:"role"`     // 角色筛选
	MemberLevel *int8  `form:"member_level"`
	Status      *int8  `form:"status"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
}

// UserListItem 用户列表项
type UserListItem struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Nickname     string     `json:"nickname"`
	Avatar       string     `json:"avatar"`
	Status       int8       `json:"status"`
	Role         int8       `json:"role"`
	MemberLevel  int8       `json:"member_level"`
	MemberExpire *time.Time `json:"member_expire"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	EmbyUserID   string     `json:"emby_user_id,omitempty"` // Emby用户ID
}

// GetUserList 获取用户列表
func (s *AdminService) GetUserList(req *UserListRequest) ([]UserListItem, int64, error) {
	query := s.db.Model(&models.User{})

	// 通用关键词搜索
	if req.Keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 精确字段筛选
	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Email != "" {
		query = query.Where("email LIKE ?", "%"+req.Email+"%")
	}
	if req.Nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+req.Nickname+"%")
	}
	if req.Role != nil {
		query = query.Where("role = ?", *req.Role)
	}
	if req.MemberLevel != nil {
		query = query.Where("member_level = ?", *req.MemberLevel)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.StartDate != "" {
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("created_at <= ?", req.EndDate+" 23:59:59")
	}

	// 统计总数
	var total int64
	query.Count(&total)

	// 分页查询
	var users []models.User
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	// 转换为列表项（直接从数据库读取emby_user_id）
	items := make([]UserListItem, len(users))
	for i, u := range users {
		items[i] = UserListItem{
			ID:           u.ID,
			Username:     u.Username,
			Email:        u.Email,
			Nickname:     u.Nickname,
			Avatar:       u.Avatar,
			Status:       u.Status,
			Role:         u.Role,
			MemberLevel:  u.MemberLevel,
			MemberExpire: u.MemberExpire,
			LastLoginAt:  u.LastLoginAt,
			CreatedAt:    u.CreatedAt,
			EmbyUserID:   u.EmbyUserID,
		}
	}

	return items, total, nil
}

// GetUserDetail 获取用户详情
func (s *AdminService) GetUserDetail(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUserStatus 更新用户状态
func (s *AdminService) UpdateUserStatus(adminID uuid.UUID, userID uuid.UUID, status int8, adminName string) error {
	if status != 1 && status != 2 {
		return errors.New("无效的状态值")
	}

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 保护超级管理员：不能禁用超级管理员
	if user.Role == models.RoleSuperAdmin {
		return errors.New("不能禁用超级管理员")
	}

	if err := s.db.Model(&user).Update("status", status).Error; err != nil {
		return err
	}

	// 同步媒体服务账号状态
	if s.isEmbyEnabled() && user.Username != "" {
		go func() {
			if adapter := s.getMediaAdapter(); adapter != nil {
				enabled := status == 1
				if err := adapter.UpdateUserStatus(user.Username, enabled); err != nil {
					log.Printf("同步媒体服务账号状态失败: %v", err)
				} else {
					action := "启用"
					if !enabled {
						action = "禁用"
					}
					log.Printf("同步媒体服务账号%s成功: %s", action, user.Username)
				}
			}
		}()
	}

	// 记录操作日志
	action := "启用用户"
	if status == 2 {
		action = "禁用用户"
	}
	s.recordOperationLog(adminID, adminName, action, userID.String(), "")

	// 清除用户缓存
	ctx := context.Background()
	database.DeleteCache(ctx, database.KeyUserInfo+userID.String())

	// 发送邮件通知
	if user.Email != "" {
		go func() {
			emailSvc := GetEmailServiceFromDB(s.db)
			if emailSvc == nil {
				return
			}
			var err error
			if status == 1 {
				err = emailSvc.SendAccountEnabled(user.Email, user.Nickname)
			} else {
				err = emailSvc.SendAccountDisabled(user.Email, user.Nickname, "管理员操作")
			}
			if err != nil {
				log.Printf("发送账户状态变更邮件失败: %v", err)
			}
		}()
	}

	return nil
}

// BatchUpdateStatus 批量更新用户状态
func (s *AdminService) BatchUpdateStatus(adminID uuid.UUID, userIDs []uuid.UUID, status int8, adminName string) error {
	if len(userIDs) == 0 {
		return errors.New("请选择要操作的用户")
	}

	// 保护超级管理员：过滤掉超级管理员
	if err := s.db.Model(&models.User{}).
		Where("id IN ?", userIDs).
		Where("role != ?", models.RoleSuperAdmin).
		Update("status", status).Error; err != nil {
		return err
	}

	// 记录操作日志
	action := "批量启用用户"
	if status == 2 {
		action = "批量禁用用户"
	}
	detail, _ := json.Marshal(userIDs)
	s.recordOperationLog(adminID, adminName, action, "", string(detail))

	// 清除缓存
	ctx := context.Background()
	for _, id := range userIDs {
		database.DeleteCache(ctx, database.KeyUserInfo+id.String())
	}

	return nil
}

// UpdateUserRole 更新用户角色
func (s *AdminService) UpdateUserRole(adminID uuid.UUID, userID uuid.UUID, role int8, adminName string) error {
	if role < 0 || role > 3 {
		return errors.New("无效的角色值")
	}

	// 获取目标用户信息
	var targetUser models.User
	if err := s.db.First(&targetUser, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 保护超级管理员：不能修改超级管理员的角色
	if targetUser.Role == models.RoleSuperAdmin {
		return errors.New("不能修改超级管理员的角色")
	}

	// 限制超级管理员数量：系统只能有一个超级管理员
	if role == models.RoleSuperAdmin {
		var count int64
		s.db.Model(&models.User{}).Where("role = ?", models.RoleSuperAdmin).Count(&count)
		if count > 0 {
			return errors.New("系统只能有一个超级管理员")
		}
	}

	// 如果升级为管理员或超级管理员，设置为长期会员
	updates := map[string]interface{}{
		"role": role,
	}
	if role >= models.RoleAdmin {
		// 设置为100年后（长期会员）
		updates["member_expire"] = time.Now().AddDate(100, 0, 0)
	}

	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return err
	}

	roleNames := []string{"普通用户", "会员用户", "管理员", "超级管理员"}
	s.recordOperationLog(adminID, adminName, "修改角色", userID.String(), "设置为: "+roleNames[role])

	return nil
}

// ResetPassword 重置用户密码
func (s *AdminService) ResetPassword(adminID uuid.UUID, userID uuid.UUID, newPassword string, adminName string) error {
	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("系统错误")
	}

	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error; err != nil {
		return err
	}

	// 同步修改媒体服务密码
	if s.isEmbyEnabled() && user.Username != "" {
		go func() {
			if adapter := s.getMediaAdapter(); adapter != nil {
				if err := adapter.UpdateUserPassword(user.Username, newPassword); err != nil {
					log.Printf("同步修改媒体服务密码失败: %v", err)
				} else {
					log.Printf("同步修改媒体服务密码成功: %s", user.Username)
				}
			}
		}()
	}

	s.recordOperationLog(adminID, adminName, "重置密码", userID.String(), "")

	// 发送邮件通知
	if user.Email != "" {
		go func() {
			emailSvc := GetEmailServiceFromDB(s.db)
			if emailSvc == nil {
				return
			}
			if err := emailSvc.SendPasswordReset(user.Email, user.Nickname, newPassword); err != nil {
				log.Printf("发送密码重置邮件失败: %v", err)
			}
		}()
	}

	return nil
}

// GetLoginLogs 获取登录日志
func (s *AdminService) GetLoginLogs(userID uuid.UUID, page, pageSize int) ([]models.LoginLog, int64, error) {
	var logs []models.LoginLog
	var total int64

	query := s.db.Model(&models.LoginLog{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetOperationLogs 获取操作日志
func (s *AdminService) GetOperationLogs(page, pageSize int, adminID *uuid.UUID) ([]models.OperationLog, int64, error) {
	var logs []models.OperationLog
	var total int64

	query := s.db.Model(&models.OperationLog{})
	if adminID != nil {
		query = query.Where("admin_id = ?", *adminID)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// recordOperationLog 记录操作日志
func (s *AdminService) recordOperationLog(adminID uuid.UUID, adminName, action, target, detail string) {
	log := &models.OperationLog{
		AdminID:   adminID,
		AdminName: adminName,
		Action:    action,
		Target:    target,
		Detail:    detail,
	}
	s.db.Create(log)
}

// UserStats 用户统计
type UserStats struct {
	TotalUsers    int64 `json:"total_users"`
	TodayRegister int64 `json:"today_register"`
	WeekRegister  int64 `json:"week_register"`
	MonthRegister int64 `json:"month_register"`
	TotalMembers  int64 `json:"total_members"`
	ActiveUsers   int64 `json:"active_users"`   // 7日内登录
	MemberConvert int64 `json:"member_convert"` // 本月会员转化
}

// GetUserStats 获取用户统计
func (s *AdminService) GetUserStats() (*UserStats, error) {
	stats := &UserStats{}
	now := time.Now()
	today := now.Format("2006-01-02")
	weekAgo := now.AddDate(0, 0, -7).Format("2006-01-02")
	monthAgo := now.AddDate(0, -1, 0).Format("2006-01-02")

	s.db.Model(&models.User{}).Count(&stats.TotalUsers)
	s.db.Model(&models.User{}).Where("DATE(created_at) = ?", today).Count(&stats.TodayRegister)
	s.db.Model(&models.User{}).Where("created_at >= ?", weekAgo).Count(&stats.WeekRegister)
	s.db.Model(&models.User{}).Where("created_at >= ?", monthAgo).Count(&stats.MonthRegister)
	s.db.Model(&models.User{}).Where("member_level > 0").Count(&stats.TotalMembers)
	s.db.Model(&models.User{}).Where("last_login_at >= ?", weekAgo).Count(&stats.ActiveUsers)

	return stats, nil
}

// DailyStats 每日统计数据
type DailyStats struct {
	Date          string `json:"date"`
	RegisterCount int64  `json:"register_count"`
	MemberCount   int64  `json:"member_count"`
}

// GetDailyStats 获取每日统计(最近30天)
func (s *AdminService) GetDailyStats(days int) ([]DailyStats, error) {
	if days <= 0 || days > 90 {
		days = 30
	}

	var results []DailyStats
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	// 注册统计
	s.db.Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as register_count
		FROM users
		WHERE created_at >= ?
		GROUP BY DATE(created_at)
		ORDER BY date
	`, startDate).Scan(&results)

	return results, nil
}

// DeleteUser 删除用户（软删除）
func (s *AdminService) DeleteUser(adminID uuid.UUID, userID uuid.UUID, adminName string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 保护超级管理员：不能删除超级管理员
	if user.Role == models.RoleSuperAdmin {
		return errors.New("不能删除超级管理员")
	}

	// 软删除用户
	if err := s.db.Delete(&user).Error; err != nil {
		return err
	}

	// 同步删除媒体服务账号
	if s.isEmbyEnabled() && user.Username != "" {
		go func() {
			if adapter := s.getMediaAdapter(); adapter != nil {
				if err := adapter.DeleteUser(user.Username); err != nil {
					log.Printf("同步删除媒体服务账号失败: %v", err)
				} else {
					log.Printf("同步删除媒体服务账号成功: %s", user.Username)
				}
			}
		}()
	}

	// 清除缓存
	ctx := context.Background()
	database.DeleteCache(ctx, database.KeyUserInfo+userID.String())
	database.DeleteCache(ctx, database.KeyUserToken+userID.String())

	// 记录操作日志
	s.recordOperationLog(adminID, adminName, "删除用户", userID.String(), user.Username)

	return nil
}

// GetEmbyUsers 获取Emby用户列表
func (s *AdminService) GetEmbyUsers() ([]emby.EmbyUserInfo, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	return adapter.GetEmbyUserList()
}

// GetEmbyUserByUsername 根据用户名获取Emby用户详情
func (s *AdminService) GetEmbyUserByUsername(username string) (*emby.EmbyUserInfo, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return nil, err
	}

	for _, eu := range embyUsers {
		if eu.Name == username {
			return &eu, nil
		}
	}

	return nil, nil // 用户不存在于Emby
}

// ========== 用户同步功能 ==========

// SyncResult 同步结果
type SyncResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	EmbyID  string `json:"emby_id,omitempty"`
}

// SyncUserToEmby 同步本地用户到Emby（在Emby创建账号）
// password 为空时使用用户名作为默认密码
func (s *AdminService) SyncUserToEmby(userID uuid.UUID, password string) (*SyncResult, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	// 获取本地用户
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	// 检查Emby是否已存在该用户
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return nil, fmt.Errorf("获取Emby用户列表失败: %v", err)
	}

	for _, eu := range embyUsers {
		if eu.Name == user.Username {
			// 用户已存在，更新本地的 emby_user_id
			if user.EmbyUserID != eu.ID {
				s.db.Model(&user).Update("emby_user_id", eu.ID)
			}
			return &SyncResult{
				Success: true,
				Message: "用户已存在于Emby",
				EmbyID:  eu.ID,
			}, nil
		}
	}

	// 使用指定密码或默认密码（用户名）
	syncPassword := password
	if syncPassword == "" {
		syncPassword = user.Username
	}

	// 在Emby创建用户
	embyID, err := adapter.CreateUser(user.Username, syncPassword)
	if err != nil {
		return nil, fmt.Errorf("创建Emby用户失败: %v", err)
	}

	// 保存 Emby 用户 ID 到本地数据库
	s.db.Model(&user).Update("emby_user_id", embyID)

	// 同步状态
	if user.Status != 1 {
		if err := adapter.UpdateUserStatus(user.Username, false); err != nil {
			log.Printf("同步用户状态失败: %v", err)
		}
	}

	return &SyncResult{
		Success: true,
		Message: "用户已同步到Emby（密码: " + syncPassword + "）",
		EmbyID:  embyID,
	}, nil
}

// SyncUserPassword 同步用户密码到Emby
func (s *AdminService) SyncUserPassword(userID uuid.UUID, password string) (*SyncResult, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	// 获取本地用户
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	// 更新Emby密码
	if err := adapter.UpdateUserPassword(user.Username, password); err != nil {
		return nil, fmt.Errorf("同步密码到Emby失败: %v", err)
	}

	return &SyncResult{
		Success: true,
		Message: "密码已同步到Emby",
	}, nil
}

// ImportEmbyUser 导入Emby用户到本地
func (s *AdminService) ImportEmbyUser(embyUsername string) (*SyncResult, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	// 检查Emby用户是否存在
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return nil, fmt.Errorf("获取Emby用户列表失败: %v", err)
	}

	var embyUser *emby.EmbyUserInfo
	for _, eu := range embyUsers {
		if eu.Name == embyUsername {
			embyUser = &eu
			break
		}
	}

	if embyUser == nil {
		return nil, errors.New("Emby用户不存在")
	}

	// 检查本地是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ?", embyUsername).First(&existingUser).Error; err == nil {
		return &SyncResult{
			Success: true,
			Message: "用户已存在于本地",
		}, nil
	}

	// 创建本地用户
	hashedPassword, _ := utils.HashPassword(embyUsername) // 默认密码与用户名相同
	status := int8(1)
	if embyUser.IsDisabled {
		status = 2
	}
	role := int8(0)
	if embyUser.IsAdmin {
		role = 2
	}

	// 生成唯一邀请码
	inviteCode := utils.GenerateInviteCode()

	newUser := &models.User{
		Username:   embyUsername,
		Nickname:   embyUsername,
		Password:   hashedPassword,
		Email:      fmt.Sprintf("%s_%s@imported.local", embyUsername, inviteCode), // 使用唯一邮箱
		Status:     status,
		Role:       role,
		InviteCode: inviteCode,
	}

	if err := s.db.Create(newUser).Error; err != nil {
		return nil, fmt.Errorf("创建本地用户失败: %v", err)
	}

	return &SyncResult{
		Success: true,
		Message: "Emby用户已导入到本地",
	}, nil
}

// SyncUserStatus 同步用户状态（修复本地与Emby状态不一致）
func (s *AdminService) SyncUserStatus(userID uuid.UUID) (*SyncResult, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	// 获取本地用户
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	// 同步状态到Emby
	enabled := user.Status == 1
	if err := adapter.UpdateUserStatus(user.Username, enabled); err != nil {
		return nil, fmt.Errorf("同步状态失败: %v", err)
	}

	action := "启用"
	if !enabled {
		action = "禁用"
	}

	return &SyncResult{
		Success: true,
		Message: fmt.Sprintf("已同步%s状态到Emby", action),
	}, nil
}

// BatchSyncResult 批量同步结果
type BatchSyncResult struct {
	Total      int      `json:"total"`
	Success    int      `json:"success"`
	Failed     int      `json:"failed"`
	Errors     []string `json:"errors,omitempty"`
	NewCreated int      `json:"new_created"` // 新创建的Emby用户数
	Imported   int      `json:"imported"`    // 导入的Emby用户数
}

// SyncAllUsers 批量同步所有用户
func (s *AdminService) SyncAllUsers() (*BatchSyncResult, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	result := &BatchSyncResult{}

	// 获取所有本地用户
	var localUsers []models.User
	if err := s.db.Find(&localUsers).Error; err != nil {
		return nil, fmt.Errorf("获取本地用户失败: %v", err)
	}

	// 获取所有Emby用户
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return nil, fmt.Errorf("获取Emby用户列表失败: %v", err)
	}

	// 建立Emby用户映射
	embyUserMap := make(map[string]*emby.EmbyUserInfo)
	for i := range embyUsers {
		embyUserMap[embyUsers[i].Name] = &embyUsers[i]
	}

	// 建立本地用户映射
	localUserMap := make(map[string]*models.User)
	for i := range localUsers {
		localUserMap[localUsers[i].Username] = &localUsers[i]
	}

	result.Total = len(localUsers)

	// 1. 同步本地用户到Emby
	for i, user := range localUsers {
		embyUser, exists := embyUserMap[user.Username]
		if !exists {
			// Emby不存在，创建
			embyID, err := adapter.CreateUser(user.Username, user.Username)
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("创建Emby用户[%s]失败: %v", user.Username, err))
				continue
			}
			// 保存 Emby 用户 ID
			if embyID != "" {
				s.db.Model(&localUsers[i]).Update("emby_user_id", embyID)
			}
			result.NewCreated++
			// 同步状态
			if user.Status != 1 {
				adapter.UpdateUserStatus(user.Username, false)
			}
			result.Success++
		} else {
			// 更新本地的 emby_user_id（如果为空）
			if user.EmbyUserID == "" && embyUser.ID != "" {
				s.db.Model(&localUsers[i]).Update("emby_user_id", embyUser.ID)
			}
			// 检查状态是否一致
			localEnabled := user.Status == 1
			embyEnabled := !embyUser.IsDisabled
			if localEnabled != embyEnabled {
				if err := adapter.UpdateUserStatus(user.Username, localEnabled); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("同步用户[%s]状态失败: %v", user.Username, err))
				}
			}
			result.Success++
		}
	}

	// 2. 导入仅存在于Emby的用户到本地（可选，这里不自动导入，只统计）
	for _, embyUser := range embyUsers {
		if _, exists := localUserMap[embyUser.Name]; !exists {
			// 仅存在于Emby，不自动导入，只记录
			result.Total++
		}
	}

	return result, nil
}

// SyncAllUsersToEmby 仅同步本地用户到Emby（不导入Emby用户）
func (s *AdminService) SyncAllUsersToEmby() (*BatchSyncResult, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	result := &BatchSyncResult{}

	// 获取所有本地用户
	var localUsers []models.User
	if err := s.db.Find(&localUsers).Error; err != nil {
		return nil, fmt.Errorf("获取本地用户失败: %v", err)
	}

	// 获取所有Emby用户
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return nil, fmt.Errorf("获取Emby用户列表失败: %v", err)
	}

	// 建立Emby用户映射
	embyUserMap := make(map[string]*emby.EmbyUserInfo)
	for i := range embyUsers {
		embyUserMap[embyUsers[i].Name] = &embyUsers[i]
	}

	result.Total = len(localUsers)

	for _, user := range localUsers {
		embyUser, exists := embyUserMap[user.Username]
		if !exists {
			// Emby不存在，创建
			_, err := adapter.CreateUser(user.Username, user.Username)
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("创建Emby用户[%s]失败: %v", user.Username, err))
				continue
			}
			result.NewCreated++
			// 同步状态
			if user.Status != 1 {
				adapter.UpdateUserStatus(user.Username, false)
			}
		} else {
			// 检查状态是否一致，同步
			localEnabled := user.Status == 1
			embyEnabled := !embyUser.IsDisabled
			if localEnabled != embyEnabled {
				if err := adapter.UpdateUserStatus(user.Username, localEnabled); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("同步用户[%s]状态失败: %v", user.Username, err))
				}
			}
		}
		result.Success++
	}

	return result, nil
}

// ImportAllEmbyUsers 导入所有仅存在于Emby的用户到本地
func (s *AdminService) ImportAllEmbyUsers() (*BatchSyncResult, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	result := &BatchSyncResult{}

	// 获取所有本地用户
	var localUsers []models.User
	if err := s.db.Find(&localUsers).Error; err != nil {
		return nil, fmt.Errorf("获取本地用户失败: %v", err)
	}

	// 获取所有Emby用户
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return nil, fmt.Errorf("获取Emby用户列表失败: %v", err)
	}

	// 建立本地用户映射
	localUserMap := make(map[string]bool)
	for _, u := range localUsers {
		localUserMap[u.Username] = true
	}

	// 导入仅存在于Emby的用户
	for _, embyUser := range embyUsers {
		if localUserMap[embyUser.Name] {
			continue // 本地已存在
		}

		result.Total++

		// 创建本地用户
		hashedPassword, _ := utils.HashPassword(embyUser.Name)
		status := int8(1)
		if embyUser.IsDisabled {
			status = 2
		}
		role := int8(0)
		if embyUser.IsAdmin {
			role = 2
		}

		// 生成唯一邀请码
		inviteCode := utils.GenerateInviteCode()

		newUser := &models.User{
			Username:   embyUser.Name,
			Nickname:   embyUser.Name,
			Password:   hashedPassword,
			Email:      fmt.Sprintf("%s_%s@imported.local", embyUser.Name, inviteCode),
			Status:     status,
			Role:       role,
			InviteCode: inviteCode,
		}

		if err := s.db.Create(newUser).Error; err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("导入用户[%s]失败: %v", embyUser.Name, err))
			continue
		}

		result.Imported++
		result.Success++
	}

	return result, nil
}

// ========== Emby设备和会话管理 ==========

// GetEmbySessions 获取所有Emby活动会话
func (s *AdminService) GetEmbySessions() ([]emby.SessionInfo, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	return adapter.GetSessions()
}

// GetEmbySessionsByUsername 根据用户名获取Emby活动会话
func (s *AdminService) GetEmbySessionsByUsername(username string) ([]emby.SessionInfo, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	// 先获取Emby用户ID
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return nil, err
	}

	var embyUserID string
	for _, u := range embyUsers {
		if u.Name == username {
			embyUserID = u.ID
			break
		}
	}

	if embyUserID == "" {
		return nil, errors.New("Emby用户不存在")
	}

	return adapter.GetSessionsByUserID(embyUserID)
}

// GetEmbyDevices 获取所有Emby注册设备
func (s *AdminService) GetEmbyDevices() ([]emby.DeviceInfo, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	return adapter.GetDevices()
}

// GetEmbyDevicesByUsername 根据用户名获取Emby设备
func (s *AdminService) GetEmbyDevicesByUsername(username string) ([]emby.DeviceInfo, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	// 先获取Emby用户ID
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return nil, err
	}

	var embyUserID string
	for _, u := range embyUsers {
		if u.Name == username {
			embyUserID = u.ID
			break
		}
	}

	if embyUserID == "" {
		return nil, errors.New("Emby用户不存在")
	}

	return adapter.GetDevicesByUserID(embyUserID)
}

// DeleteEmbyDevice 删除Emby设备
func (s *AdminService) DeleteEmbyDevice(deviceID string) error {
	if !s.isEmbyEnabled() {
		return errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return errors.New("无法连接媒体服务")
	}

	return adapter.DeleteDevice(deviceID)
}

// KillEmbySession 终止Emby会话（强制下线）
func (s *AdminService) KillEmbySession(sessionID string) error {
	if !s.isEmbyEnabled() {
		return errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return errors.New("无法连接媒体服务")
	}

	return adapter.KillSession(sessionID)
}

// CheckAndEnforceSessionLimit 检查并强制执行会话限制
// 返回被终止的会话数量
func (s *AdminService) CheckAndEnforceSessionLimit() (int, error) {
	if !s.isEmbyEnabled() {
		return 0, errors.New("媒体服务未启用")
	}

	// 获取会话限制设置
	settingService := NewSettingService(s.db)
	limitSettings, err := settingService.GetSessionLimitSettings()
	if err != nil {
		return 0, err
	}

	// 如果未启用限制，直接返回
	if !limitSettings.Enabled || limitSettings.MaxSessions <= 0 {
		return 0, nil
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return 0, errors.New("无法连接媒体服务")
	}

	// 获取所有会话
	sessions, err := adapter.GetSessions()
	if err != nil {
		return 0, err
	}

	// 按用户分组会话
	userSessions := make(map[string][]emby.SessionInfo)
	for _, session := range sessions {
		if session.UserName != "" {
			userSessions[session.UserName] = append(userSessions[session.UserName], session)
		}
	}

	killedCount := 0

	// 检查每个用户的会话数
	for userName, userSessionList := range userSessions {
		if len(userSessionList) > limitSettings.MaxSessions {
			// 超出限制，需要终止多余的会话
			excessCount := len(userSessionList) - limitSettings.MaxSessions

			if limitSettings.AutoKillOldest {
				// 按最后活动时间排序，终止最早的会话
				// 排序：最早的在前面
				sortedSessions := make([]emby.SessionInfo, len(userSessionList))
				copy(sortedSessions, userSessionList)
				
				// 简单冒泡排序按时间升序
				for i := 0; i < len(sortedSessions)-1; i++ {
					for j := 0; j < len(sortedSessions)-i-1; j++ {
						if sortedSessions[j].LastActivityDate > sortedSessions[j+1].LastActivityDate {
							sortedSessions[j], sortedSessions[j+1] = sortedSessions[j+1], sortedSessions[j]
						}
					}
				}

				// 终止最早的会话
				for i := 0; i < excessCount && i < len(sortedSessions); i++ {
					if err := adapter.KillSession(sortedSessions[i].ID); err != nil {
						log.Printf("终止用户 %s 的会话 %s 失败: %v", userName, sortedSessions[i].ID, err)
					} else {
						log.Printf("已终止用户 %s 的会话 %s (设备: %s)", userName, sortedSessions[i].ID, sortedSessions[i].DeviceName)
						killedCount++
					}
				}
			}
		}
	}

	return killedCount, nil
}

// GetSessionLimitStatus 获取会话限制状态
func (s *AdminService) GetSessionLimitStatus() (map[string]interface{}, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	// 获取会话限制设置
	settingService := NewSettingService(s.db)
	limitSettings, err := settingService.GetSessionLimitSettings()
	if err != nil {
		return nil, err
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	// 获取所有会话
	sessions, err := adapter.GetSessions()
	if err != nil {
		return nil, err
	}

	// 按用户分组统计
	userSessionCounts := make(map[string]int)
	usersOverLimit := make([]string, 0)
	
	for _, session := range sessions {
		if session.UserName != "" {
			userSessionCounts[session.UserName]++
		}
	}

	// 检查超限用户
	if limitSettings.Enabled && limitSettings.MaxSessions > 0 {
		for userName, count := range userSessionCounts {
			if count > limitSettings.MaxSessions {
				usersOverLimit = append(usersOverLimit, userName)
			}
		}
	}

	// 统计播放中的会话
	userPlayingCounts := make(map[string]int)
	usersOverPlayLimit := make([]string, 0)
	totalPlaying := 0

	for _, session := range sessions {
		if session.UserName != "" && session.IsPlaying {
			userPlayingCounts[session.UserName]++
			totalPlaying++
		}
	}

	// 检查播放超限用户
	if limitSettings.PlayLimitEnabled && limitSettings.MaxPlayingSessions > 0 {
		for userName, count := range userPlayingCounts {
			if count > limitSettings.MaxPlayingSessions {
				usersOverPlayLimit = append(usersOverPlayLimit, userName)
			}
		}
	}

	return map[string]interface{}{
		"enabled":               limitSettings.Enabled,
		"max_sessions":          limitSettings.MaxSessions,
		"auto_kill_oldest":      limitSettings.AutoKillOldest,
		"play_limit_enabled":    limitSettings.PlayLimitEnabled,
		"max_playing_sessions":  limitSettings.MaxPlayingSessions,
		"auto_stop_oldest_play": limitSettings.AutoStopOldestPlay,
		"total_sessions":        len(sessions),
		"total_playing":         totalPlaying,
		"unique_users":          len(userSessionCounts),
		"users_over_limit":      usersOverLimit,
		"users_over_play_limit": usersOverPlayLimit,
	}, nil
}

// CheckAndEnforcePlayLimit 检查并强制执行播放数量限制
// 返回被停止的播放数量
func (s *AdminService) CheckAndEnforcePlayLimit() (int, error) {
	if !s.isEmbyEnabled() {
		return 0, errors.New("媒体服务未启用")
	}

	// 获取会话限制设置
	settingService := NewSettingService(s.db)
	limitSettings, err := settingService.GetSessionLimitSettings()
	if err != nil {
		return 0, err
	}

	// 如果未启用播放限制，直接返回
	if !limitSettings.PlayLimitEnabled || limitSettings.MaxPlayingSessions <= 0 {
		return 0, nil
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return 0, errors.New("无法连接媒体服务")
	}

	// 获取所有会话
	sessions, err := adapter.GetSessions()
	if err != nil {
		return 0, err
	}

	// 按用户分组正在播放的会话
	userPlayingSessions := make(map[string][]emby.SessionInfo)
	for _, session := range sessions {
		if session.UserName != "" && session.IsPlaying {
			userPlayingSessions[session.UserName] = append(userPlayingSessions[session.UserName], session)
		}
	}

	stoppedCount := 0

	// 检查每个用户的播放数
	for userName, playingList := range userPlayingSessions {
		if len(playingList) > limitSettings.MaxPlayingSessions {
			// 超出限制，需要停止多余的播放
			excessCount := len(playingList) - limitSettings.MaxPlayingSessions

			if limitSettings.AutoStopOldestPlay {
				// 按最后活动时间排序，停止最早的播放
				sortedSessions := make([]emby.SessionInfo, len(playingList))
				copy(sortedSessions, playingList)

				// 简单冒泡排序按时间升序（最早的在前面）
				for i := 0; i < len(sortedSessions)-1; i++ {
					for j := 0; j < len(sortedSessions)-i-1; j++ {
						if sortedSessions[j].LastActivityDate > sortedSessions[j+1].LastActivityDate {
							sortedSessions[j], sortedSessions[j+1] = sortedSessions[j+1], sortedSessions[j]
						}
					}
				}

				// 停止最早的播放
				for i := 0; i < excessCount && i < len(sortedSessions); i++ {
					if err := adapter.StopPlayback(sortedSessions[i].ID); err != nil {
						log.Printf("停止用户 %s 的播放会话 %s 失败: %v", userName, sortedSessions[i].ID, err)
					} else {
						itemName := ""
						if sortedSessions[i].NowPlayingItem.Name != "" {
							itemName = sortedSessions[i].NowPlayingItem.Name
						}
						log.Printf("已停止用户 %s 的播放会话 %s (设备: %s, 正在播放: %s)",
							userName, sortedSessions[i].ID, sortedSessions[i].DeviceName, itemName)
						stoppedCount++
					}
				}
			}
		}
	}

	return stoppedCount, nil
}

// SetEmbyUserStreamLimit 设置Emby用户同时在线流数限制
func (s *AdminService) SetEmbyUserStreamLimit(username string, limit int) error {
	if !s.isEmbyEnabled() {
		return errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return errors.New("无法连接媒体服务")
	}

	// 先获取Emby用户ID
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return err
	}

	var embyUserID string
	for _, u := range embyUsers {
		if u.Name == username {
			embyUserID = u.ID
			break
		}
	}

	if embyUserID == "" {
		return errors.New("Emby用户不存在")
	}

	return adapter.SetUserStreamLimit(embyUserID, limit)
}

// GetEmbyUserStreamLimit 获取Emby用户同时在线流数限制
func (s *AdminService) GetEmbyUserStreamLimit(username string) (int, error) {
	if !s.isEmbyEnabled() {
		return 0, errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return 0, errors.New("无法连接媒体服务")
	}

	// 先获取Emby用户ID
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return 0, err
	}

	var embyUserID string
	for _, u := range embyUsers {
		if u.Name == username {
			embyUserID = u.ID
			break
		}
	}

	if embyUserID == "" {
		return 0, errors.New("Emby用户不存在")
	}

	return adapter.GetUserStreamLimit(embyUserID)
}

// ========== 客户端白名单强制执行 ==========

// EnforceClientWhitelistResult 强制执行结果
type EnforceClientWhitelistResult struct {
	TotalSessions   int      `json:"total_sessions"`
	KilledSessions  int      `json:"killed_sessions"`
	AllowedClients  []string `json:"allowed_clients"`
	BlockedClients  []string `json:"blocked_clients"`
}

// EnforceClientWhitelist 强制执行客户端白名单（踢出未授权客户端）
func (s *AdminService) EnforceClientWhitelist() (*EnforceClientWhitelistResult, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	// 获取客户端白名单设置
	settingService := NewSettingService(s.db)
	whitelistSettings, err := settingService.GetClientWhitelistSettings()
	if err != nil {
		return nil, fmt.Errorf("获取客户端白名单设置失败: %v", err)
	}

	// 如果白名单未启用，不执行任何操作
	if !whitelistSettings.Enabled {
		return &EnforceClientWhitelistResult{
			TotalSessions:  0,
			KilledSessions: 0,
		}, nil
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	// 获取所有会话
	sessions, err := adapter.GetSessions()
	if err != nil {
		return nil, fmt.Errorf("获取会话列表失败: %v", err)
	}

	result := &EnforceClientWhitelistResult{
		TotalSessions:  len(sessions),
		AllowedClients: make([]string, 0),
		BlockedClients: make([]string, 0),
	}

	// 检查每个会话的客户端
	for _, session := range sessions {
		clientName := session.Client
		if clientName == "" {
			continue
		}

		// 检查客户端是否在白名单中
		allowed := settingService.IsClientAllowed(clientName)
		
		if allowed {
			// 记录允许的客户端（去重）
			found := false
			for _, c := range result.AllowedClients {
				if c == clientName {
					found = true
					break
				}
			}
			if !found {
				result.AllowedClients = append(result.AllowedClients, clientName)
			}
		} else {
			// 踢出未授权客户端
			if err := adapter.KillSession(session.ID); err != nil {
				log.Printf("踢出会话失败 [%s - %s]: %v", session.UserName, clientName, err)
			} else {
				result.KilledSessions++
				log.Printf("已踢出未授权客户端: 用户=%s, 客户端=%s, 设备=%s", 
					session.UserName, clientName, session.DeviceName)
			}
			
			// 记录被阻止的客户端（去重）
			found := false
			for _, c := range result.BlockedClients {
				if c == clientName {
					found = true
					break
				}
			}
			if !found {
				result.BlockedClients = append(result.BlockedClients, clientName)
			}
		}
	}

	return result, nil
}


// ========== 用户设备策略管理（使用Emby的EnableAllDevices/EnabledDevices） ==========

// UserDevicePolicy 用户设备策略
type UserDevicePolicy struct {
	EnableAllDevices bool     `json:"enable_all_devices"`
	EnabledDevices   []string `json:"enabled_devices"`   // 已授权的设备ID列表
	EnabledClients   []string `json:"enabled_clients"`   // 已授权的客户端名称列表（如 VidHub, SenPlayer）
}

// GetEmbyUserDevicePolicy 获取用户设备策略
func (s *AdminService) GetEmbyUserDevicePolicy(username string) (*UserDevicePolicy, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	// 先获取Emby用户ID
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return nil, err
	}

	var embyUserID string
	for _, u := range embyUsers {
		if u.Name == username {
			embyUserID = u.ID
			break
		}
	}

	if embyUserID == "" {
		return nil, errors.New("Emby用户不存在")
	}

	enableAll, enabledDevices, err := adapter.GetUserDevicePolicy(embyUserID)
	if err != nil {
		return nil, err
	}

	// 获取所有设备，根据已授权的设备ID反推客户端名称
	allDevices, _ := adapter.GetDevices()
	enabledClientsMap := make(map[string]bool)
	for _, deviceID := range enabledDevices {
		for _, device := range allDevices {
			if device.ID == deviceID && device.AppName != "" {
				enabledClientsMap[device.AppName] = true
				break
			}
		}
	}
	
	enabledClients := make([]string, 0, len(enabledClientsMap))
	for client := range enabledClientsMap {
		enabledClients = append(enabledClients, client)
	}

	return &UserDevicePolicy{
		EnableAllDevices: enableAll,
		EnabledDevices:   enabledDevices,
		EnabledClients:   enabledClients,
	}, nil
}

// SetEmbyUserDevicePolicy 设置用户设备策略
func (s *AdminService) SetEmbyUserDevicePolicy(username string, enableAll bool, deviceIDs []string) error {
	if !s.isEmbyEnabled() {
		return errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return errors.New("无法连接媒体服务")
	}

	// 先获取Emby用户ID
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return err
	}

	var embyUserID string
	for _, u := range embyUsers {
		if u.Name == username {
			embyUserID = u.ID
			break
		}
	}

	if embyUserID == "" {
		return errors.New("Emby用户不存在")
	}

	return adapter.SetUserDevicePolicy(embyUserID, enableAll, deviceIDs)
}

// AddDeviceToEmbyUserWhitelist 添加设备到用户白名单
func (s *AdminService) AddDeviceToEmbyUserWhitelist(username, deviceID string) error {
	if !s.isEmbyEnabled() {
		return errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return errors.New("无法连接媒体服务")
	}

	// 先获取Emby用户ID
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return err
	}

	var embyUserID string
	for _, u := range embyUsers {
		if u.Name == username {
			embyUserID = u.ID
			break
		}
	}

	if embyUserID == "" {
		return errors.New("Emby用户不存在")
	}

	return adapter.AddDeviceToUserWhitelist(embyUserID, deviceID)
}

// RemoveDeviceFromEmbyUserWhitelist 从用户白名单移除设备
func (s *AdminService) RemoveDeviceFromEmbyUserWhitelist(username, deviceID string) error {
	if !s.isEmbyEnabled() {
		return errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return errors.New("无法连接媒体服务")
	}

	// 先获取Emby用户ID
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return err
	}

	var embyUserID string
	for _, u := range embyUsers {
		if u.Name == username {
			embyUserID = u.ID
			break
		}
	}

	if embyUserID == "" {
		return errors.New("Emby用户不存在")
	}

	return adapter.RemoveDeviceFromUserWhitelist(embyUserID, deviceID)
}

// SetEmbyUserClientPolicy 按客户端名称设置用户设备策略
// clientNames: 允许的客户端名称列表，如 ["VidHub", "SenPlayer", "Emby Web"]
func (s *AdminService) SetEmbyUserClientPolicy(username string, enableAll bool, clientNames []string) error {
	if !s.isEmbyEnabled() {
		return errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return errors.New("无法连接媒体服务")
	}

	// 先获取Emby用户ID
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return err
	}

	var embyUserID string
	for _, u := range embyUsers {
		if u.Name == username {
			embyUserID = u.ID
			break
		}
	}

	if embyUserID == "" {
		return errors.New("Emby用户不存在")
	}

	if enableAll {
		// 允许所有设备
		return adapter.SetUserDevicePolicy(embyUserID, true, nil)
	}

	// 获取所有设备，找出匹配客户端名称的设备ID
	allDevices, err := adapter.GetDevices()
	if err != nil {
		return err
	}

	// 创建客户端名称集合用于快速查找
	clientSet := make(map[string]bool)
	for _, name := range clientNames {
		clientSet[name] = true
	}

	// 收集匹配的设备ID
	var deviceIDs []string
	for _, device := range allDevices {
		if clientSet[device.AppName] {
			deviceIDs = append(deviceIDs, device.ID)
		}
	}

	return adapter.SetUserDevicePolicy(embyUserID, false, deviceIDs)
}

// ApplyGlobalDeviceWhitelistToUser 将全局客户端白名单应用到用户
// 这会根据全局白名单设置用户的EnableAllDevices和EnabledDevices
func (s *AdminService) ApplyGlobalDeviceWhitelistToUser(username string) error {
	if !s.isEmbyEnabled() {
		return errors.New("媒体服务未启用")
	}

	// 获取全局客户端白名单设置
	settingService := NewSettingService(s.db)
	whitelistSettings, err := settingService.GetClientWhitelistSettings()
	if err != nil {
		return fmt.Errorf("获取客户端白名单设置失败: %v", err)
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return errors.New("无法连接媒体服务")
	}

	// 先获取Emby用户ID
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return err
	}

	var embyUserID string
	for _, u := range embyUsers {
		if u.Name == username {
			embyUserID = u.ID
			break
		}
	}

	if embyUserID == "" {
		return errors.New("Emby用户不存在")
	}

	// 如果全局白名单未启用，允许所有设备
	if !whitelistSettings.Enabled {
		return adapter.SetUserDevicePolicy(embyUserID, true, []string{})
	}

	// 获取用户当前使用的设备
	devices, err := adapter.GetDevicesByUserID(embyUserID)
	if err != nil {
		return fmt.Errorf("获取用户设备失败: %v", err)
	}

	// 筛选出在白名单中的设备
	var allowedDeviceIDs []string
	for _, device := range devices {
		if settingService.IsClientAllowed(device.AppName) {
			allowedDeviceIDs = append(allowedDeviceIDs, device.ID)
		}
	}

	// 设置用户设备策略：禁用所有设备，只允许白名单中的设备
	return adapter.SetUserDevicePolicy(embyUserID, false, allowedDeviceIDs)
}

// ApplyGlobalDeviceWhitelistToAllUsers 将全局客户端白名单应用到所有用户
func (s *AdminService) ApplyGlobalDeviceWhitelistToAllUsers() (*BatchSyncResult, error) {
	if !s.isEmbyEnabled() {
		return nil, errors.New("媒体服务未启用")
	}

	adapter := s.getMediaAdapter()
	if adapter == nil {
		return nil, errors.New("无法连接媒体服务")
	}

	result := &BatchSyncResult{}

	// 获取所有Emby用户
	embyUsers, err := adapter.GetEmbyUserList()
	if err != nil {
		return nil, fmt.Errorf("获取Emby用户列表失败: %v", err)
	}

	result.Total = len(embyUsers)

	for _, user := range embyUsers {
		// 跳过管理员
		if user.IsAdmin {
			result.Success++
			continue
		}

		if err := s.ApplyGlobalDeviceWhitelistToUser(user.Name); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("用户[%s]: %v", user.Name, err))
		} else {
			result.Success++
		}
	}

	return result, nil
}
