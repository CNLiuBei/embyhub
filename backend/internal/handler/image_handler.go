package handler

import (
	"io"
	"log"
	"net/http"
	"strings"

	"feiniu-user-system/internal/database"
	"feiniu-user-system/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ImageHandler 图片处理器
type ImageHandler struct {
	db *gorm.DB
}

// NewImageHandler 创建图片处理器
func NewImageHandler(db *gorm.DB) *ImageHandler {
	return &ImageHandler{db: db}
}

// getEmbySettings 获取Emby设置
func (h *ImageHandler) getEmbySettings() *service.EmbySettings {
	settingService := service.NewSettingService(h.db)
	settings, err := settingService.GetEmbySettings()
	if err != nil {
		return nil
	}
	return settings
}

// ProxyImage 代理媒体服务图片
// @Summary 代理媒体服务图片
// @Tags 媒体
// @Security Bearer
// @Param path path string true "图片路径"
// @Param uid query string true "用户ID"
// @Success 200 "图片内容"
// @Router /api/v1/image/{path} [get]
func (h *ImageHandler) ProxyImage(c *gin.Context) {
	// 获取Emby设置
	settings := h.getEmbySettings()
	if settings == nil || !settings.Enabled {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	// 获取图片路径 - Gin的*path参数会包含前导斜杠
	imagePath := c.Param("path")
	
	// 获取原始查询参数（可能包含maxWidth等）
	rawQuery := c.Request.URL.RawQuery

	var imageURL string
	var req *http.Request
	var err error

	// 根据模式构建不同的图片URL
	if settings.IsEmbyMode() {
		// Emby模式 - 图片路径可能是完整URL或相对路径
		if strings.HasPrefix(imagePath, "/http://") || strings.HasPrefix(imagePath, "/https://") {
			// 去掉前导斜杠
			imageURL = imagePath[1:]
		} else if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
			imageURL = imagePath
		} else {
			// Emby图片URL格式: /Items/{itemId}/Images/{imageType}
			imageURL = settings.BaseURL + imagePath
		}
		
		// 处理查询参数 - 提取maxWidth等参数，忽略uid
		if rawQuery != "" {
			// 解析查询参数，过滤掉uid
			params := strings.Split(rawQuery, "&")
			var filteredParams []string
			for _, p := range params {
				if !strings.HasPrefix(p, "uid=") {
					filteredParams = append(filteredParams, p)
				}
			}
			if len(filteredParams) > 0 {
				cleanQuery := strings.Join(filteredParams, "&")
				if strings.Contains(imageURL, "?") {
					imageURL += "&" + cleanQuery
				} else {
					imageURL += "?" + cleanQuery
				}
			}
		}

		req, err = http.NewRequest("GET", imageURL, nil)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		// Emby使用API Key认证
		if settings.APIKey != "" {
			req.Header.Set("X-Emby-Token", settings.APIKey)
		}
	} else {
		// 飞牛模式
		userID := c.Query("uid")
		if userID == "" {
			c.Status(http.StatusBadRequest)
			return
		}

		// 获取用户的飞牛token
		ctx := c.Request.Context()
		embyKey := "emby:token:" + userID
		token, err := database.GetCache(ctx, embyKey)

		if err != nil || token == "" {
			c.Status(http.StatusUnauthorized)
			return
		}

		// 构建飞牛图片URL
		imageURL = settings.BaseURL + "/sys/img" + imagePath

		req, err = http.NewRequest("GET", imageURL, nil)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		// 飞牛使用token认证
		req.Header.Set("authorization", token)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		// 调试日志
		log.Printf("图片代理失败: URL=%s, Status=%d", imageURL, resp.StatusCode)
		c.Status(resp.StatusCode)
		return
	}

	// 设置响应头和状态码
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=86400") // 缓存24小时
	c.Status(http.StatusOK)

	// 复制响应体
	io.Copy(c.Writer, resp.Body)
}
