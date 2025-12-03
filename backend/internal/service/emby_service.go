package service

import (
	"embyhub/config"
	"embyhub/internal/dao"
	"embyhub/internal/model"
	"embyhub/pkg/emby"
	"fmt"
)

type EmbyService struct {
	client  *emby.Client
	userDAO *dao.UserDAO
}

func NewEmbyService() *EmbyService {
	client := emby.NewClient(&config.GlobalConfig.Emby)
	return &EmbyService{
		client:  client,
		userDAO: dao.NewUserDAO(),
	}
}

// TestConnection 测试Emby连接
func (s *EmbyService) TestConnection() error {
	return s.client.TestConnection()
}

// SyncUsers 同步Emby用户到本地系统
func (s *EmbyService) SyncUsers() (int, error) {
	// 获取Emby用户列表
	embyUsers, err := s.client.GetUsers()
	if err != nil {
		return 0, fmt.Errorf("获取Emby用户失败: %w", err)
	}

	syncCount := 0

	// 同步到本地数据库
	for _, embyUser := range embyUsers {
		// 1. 先检查是否已通过emby_user_id关联
		existingUser, _ := s.userDAO.GetByEmbyUserID(embyUser.ID)
		if existingUser != nil {
			// 已关联，跳过
			continue
		}

		// 2. 检查是否有同名用户
		existingUser, _ = s.userDAO.GetByUsername(embyUser.Name)
		if existingUser != nil {
			// 存在同名用户，关联emby_user_id
			if existingUser.EmbyUserID == "" {
				existingUser.EmbyUserID = embyUser.ID
				if err := s.userDAO.Update(existingUser); err != nil {
					continue
				}
				syncCount++
			}
			continue
		}

		// 3. 创建新用户（使用默认密码，角色为普通用户）
		// 生成唯一邮箱避免约束冲突
		email := fmt.Sprintf("%s@emby.sync", embyUser.Name)
		newUser := &model.User{
			Username:     embyUser.Name,
			PasswordHash: "$2a$10$defaulthashforsyncedusersneedtoreset", // 需要重置密码
			Email:        email,
			EmbyUserID:   embyUser.ID,
			RoleID:       3, // 普通用户
			Status:       1,
		}
		if err := s.userDAO.Create(newUser); err != nil {
			fmt.Printf("创建同步用户失败: %s - %v\n", embyUser.Name, err)
			continue
		}
		syncCount++
	}

	return syncCount, nil
}

// GetUsers 获取Emby用户列表
func (s *EmbyService) GetUsers() ([]*model.EmbyUser, error) {
	return s.client.GetUsers()
}

// ========== 媒体库相关方法 ==========

// GetLibraries 获取媒体库列表
func (s *EmbyService) GetLibraries(embyUserId string) ([]emby.MediaLibrary, error) {
	return s.client.GetLibraries(embyUserId)
}

// GetItems 获取媒体项目列表
func (s *EmbyService) GetItems(parentId string, itemType string, startIndex, limit int, sortBy, sortOrder, searchTerm string) (*emby.MediaItemsResponse, error) {
	return s.client.GetItems(parentId, itemType, startIndex, limit, sortBy, sortOrder, searchTerm)
}

// GetItem 获取单个媒体详情
func (s *EmbyService) GetItem(itemId string) (*emby.MediaItem, error) {
	return s.client.GetItem(itemId)
}

// GetLatestItems 获取最新媒体
func (s *EmbyService) GetLatestItems(embyUserId string, parentId string, limit int) ([]emby.MediaItem, error) {
	return s.client.GetLatestItems(embyUserId, parentId, limit)
}

// GetImageURL 获取媒体图片URL
func (s *EmbyService) GetImageURL(itemId string, imageType string, tag string) string {
	return s.client.GetImageURL(itemId, imageType, tag)
}

// GetServerURL 获取Emby服务器URL
func (s *EmbyService) GetServerURL() string {
	return s.client.ServerURL
}
