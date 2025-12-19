// Package handler 壁纸处理器
package handler

import (
	"net/http"
	"os"

	"feiniu-user-system/internal/service"

	"github.com/gin-gonic/gin"
)

// WallpaperHandler 壁纸处理器
type WallpaperHandler struct {
	wallpaperService *service.WallpaperService
}

// NewWallpaperHandler 创建壁纸处理器
func NewWallpaperHandler(ws *service.WallpaperService) *WallpaperHandler {
	return &WallpaperHandler{
		wallpaperService: ws,
	}
}

// GetWallpaper 获取当前壁纸
func (h *WallpaperHandler) GetWallpaper(c *gin.Context) {
	path, err := h.wallpaperService.GetWallpaperPath()
	if err != nil {
		// 如果没有本地壁纸，重定向到Bing
		c.Redirect(http.StatusTemporaryRedirect, "https://bing.img.run/1920x1080.php")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		c.Redirect(http.StatusTemporaryRedirect, "https://bing.img.run/1920x1080.php")
		return
	}

	// 设置缓存头
	c.Header("Cache-Control", "public, max-age=3600")
	c.File(path)
}
