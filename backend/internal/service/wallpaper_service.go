// Package service 壁纸服务
package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// BingWallpaperResponse Bing壁纸API响应
type BingWallpaperResponse struct {
	Images []struct {
		URL       string `json:"url"`
		URLBase   string `json:"urlbase"`
		Copyright string `json:"copyright"`
		StartDate string `json:"startdate"`
		EndDate   string `json:"enddate"`
		Hsh       string `json:"hsh"`
	} `json:"images"`
}

// WallpaperService 壁纸服务
type WallpaperService struct {
	logger       *zap.Logger
	wallpaperDir string
	currentHash  string
	mu           sync.RWMutex
}

// NewWallpaperService 创建壁纸服务
func NewWallpaperService(logger *zap.Logger) *WallpaperService {
	wallpaperDir := "./static/wallpaper"

	// 确保目录存在
	if err := os.MkdirAll(wallpaperDir, 0755); err != nil {
		logger.Error("创建壁纸目录失败", zap.Error(err))
	}

	ws := &WallpaperService{
		logger:       logger,
		wallpaperDir: wallpaperDir,
	}

	// 启动时检查并下载壁纸
	go ws.CheckAndDownload()

	// 启动定时任务，每小时检查一次
	go ws.startScheduler()

	return ws
}

// startScheduler 启动定时调度器
// 每天凌晨2点更新壁纸
func (ws *WallpaperService) startScheduler() {
	for {
		now := time.Now()
		// 计算下一个2点的时间
		next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
		if now.After(next) {
			// 如果今天2点已过，则设置为明天2点
			next = next.Add(24 * time.Hour)
		}

		duration := time.Until(next)
		ws.logger.Info("下次壁纸更新时间", zap.Time("next", next), zap.Duration("duration", duration))

		time.Sleep(duration)
		ws.CheckAndDownload()
	}
}

// CheckAndDownload 检查并下载壁纸
func (ws *WallpaperService) CheckAndDownload() {
	ws.logger.Info("开始检查Bing壁纸更新")

	// 获取Bing壁纸信息
	resp, err := http.Get("https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1&mkt=zh-CN")
	if err != nil {
		ws.logger.Error("获取Bing壁纸信息失败", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	var bingResp BingWallpaperResponse
	if err := json.NewDecoder(resp.Body).Decode(&bingResp); err != nil {
		ws.logger.Error("解析Bing壁纸响应失败", zap.Error(err))
		return
	}

	if len(bingResp.Images) == 0 {
		ws.logger.Warn("Bing壁纸响应中没有图片")
		return
	}

	image := bingResp.Images[0]
	newHash := image.Hsh

	ws.mu.RLock()
	currentHash := ws.currentHash
	ws.mu.RUnlock()

	// 检查是否需要更新
	if currentHash == newHash {
		ws.logger.Info("壁纸无更新")
		return
	}

	// 下载原图 (UHD分辨率)
	imageURL := fmt.Sprintf("https://www.bing.com%s_UHD.jpg", image.URLBase)
	ws.logger.Info("下载Bing壁纸", zap.String("url", imageURL))

	imgResp, err := http.Get(imageURL)
	if err != nil {
		ws.logger.Error("下载壁纸失败", zap.Error(err))
		return
	}
	defer imgResp.Body.Close()

	if imgResp.StatusCode != http.StatusOK {
		ws.logger.Error("下载壁纸失败", zap.Int("status", imgResp.StatusCode))
		return
	}

	// 删除旧壁纸
	ws.cleanOldWallpapers()

	// 保存新壁纸
	filename := fmt.Sprintf("bing_%s.jpg", image.StartDate)
	filepath := filepath.Join(ws.wallpaperDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		ws.logger.Error("创建壁纸文件失败", zap.Error(err))
		return
	}
	defer file.Close()

	written, err := io.Copy(file, imgResp.Body)
	if err != nil {
		ws.logger.Error("保存壁纸失败", zap.Error(err))
		return
	}

	ws.mu.Lock()
	ws.currentHash = newHash
	ws.mu.Unlock()

	ws.logger.Info("壁纸下载成功",
		zap.String("filename", filename),
		zap.Int64("size", written),
		zap.String("copyright", image.Copyright),
	)
}

// cleanOldWallpapers 清理旧壁纸
func (ws *WallpaperService) cleanOldWallpapers() {
	files, err := os.ReadDir(ws.wallpaperDir)
	if err != nil {
		ws.logger.Error("读取壁纸目录失败", zap.Error(err))
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := filepath.Join(ws.wallpaperDir, file.Name())
		if err := os.Remove(filePath); err != nil {
			ws.logger.Error("删除旧壁纸失败", zap.String("file", file.Name()), zap.Error(err))
		} else {
			ws.logger.Info("已删除旧壁纸", zap.String("file", file.Name()))
		}
	}
}

// GetWallpaperPath 获取当前壁纸路径
func (ws *WallpaperService) GetWallpaperPath() (string, error) {
	files, err := os.ReadDir(ws.wallpaperDir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".jpg" {
			return filepath.Join(ws.wallpaperDir, file.Name()), nil
		}
	}

	return "", fmt.Errorf("no wallpaper found")
}
