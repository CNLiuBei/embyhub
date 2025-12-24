// Package service Cloudflare Tunnel 服务
package service

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// CloudflareTunnelConfig 隧道配置
type CloudflareTunnelConfig struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	TunnelID      string    `json:"tunnel_id" gorm:"size:100"`           // Cloudflare 隧道 ID
	TunnelName    string    `json:"tunnel_name" gorm:"size:100"`         // 隧道名称
	Domain        string    `json:"domain" gorm:"size:200"`              // 主域名 (如 liubei.org)
	Subdomain     string    `json:"subdomain" gorm:"size:100"`           // 子域名前缀 (如 pay)
	FullDomain    string    `json:"full_domain" gorm:"size:300"`         // 完整域名 (如 pay.liubei.org)
	LocalPort     int       `json:"local_port" gorm:"default:54680"`     // 本地端口
	Status        string    `json:"status" gorm:"size:20;default:stopped"` // running, stopped, error
	ConfigPath    string    `json:"config_path" gorm:"size:500"`         // 配置文件路径
	CredPath      string    `json:"cred_path" gorm:"size:500"`           // 凭证文件路径
	ErrorMsg      string    `json:"error_msg" gorm:"size:1000"`          // 错误信息
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 表名
func (CloudflareTunnelConfig) TableName() string {
	return "cloudflare_tunnel_configs"
}

// CloudflareTunnelService 隧道服务
type CloudflareTunnelService struct {
	db          *gorm.DB
	manager     *CloudflaredManager // cloudflared 二进制管理器
	tunnelCmd   *exec.Cmd
	tunnelMutex sync.Mutex
	cancelFunc  context.CancelFunc
}

// NewCloudflareTunnelService 创建隧道服务
func NewCloudflareTunnelService(db *gorm.DB) *CloudflareTunnelService {
	// 自动迁移表
	db.AutoMigrate(&CloudflareTunnelConfig{})
	
	return &CloudflareTunnelService{
		db:      db,
		manager: NewCloudflaredManager(),
	}
}

// GetConfig 获取隧道配置
func (s *CloudflareTunnelService) GetConfig() (*CloudflareTunnelConfig, error) {
	var config CloudflareTunnelConfig
	err := s.db.First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &CloudflareTunnelConfig{
				LocalPort: 54680,
				Status:    "stopped",
			}, nil
		}
		return nil, err
	}
	return &config, nil
}

// CheckCloudflaredInstalled 检查 cloudflared 是否安装（优先检查项目内置的）
func (s *CloudflareTunnelService) CheckCloudflaredInstalled() (bool, string) {
	// 优先检查项目内置的 cloudflared
	if s.manager.IsInstalled() {
		version, err := s.manager.GetVersion()
		if err == nil {
			return true, version
		}
	}
	
	// 检查系统安装的 cloudflared
	cmd := exec.Command("cloudflared", "--version")
	output, err := cmd.Output()
	if err != nil {
		return false, ""
	}
	return true, strings.TrimSpace(string(output))
}

// getCloudflaredPath 获取 cloudflared 可执行文件路径
func (s *CloudflareTunnelService) getCloudflaredPath() string {
	// 优先使用项目内置的
	if s.manager.IsInstalled() {
		return s.manager.GetBinPath()
	}
	// 回退到系统安装的
	return "cloudflared"
}

// CheckLoginStatus 检查是否已登录 Cloudflare
func (s *CloudflareTunnelService) CheckLoginStatus() bool {
	return s.manager.IsLoggedIn()
}

// DownloadCloudflared 下载 cloudflared 二进制文件
func (s *CloudflareTunnelService) DownloadCloudflared(progressCallback func(downloaded, total int64)) error {
	return s.manager.Download(progressCallback)
}

// IsCloudflaredDownloaded 检查是否已下载内置的 cloudflared
func (s *CloudflareTunnelService) IsCloudflaredDownloaded() bool {
	return s.manager.IsInstalled()
}

// GetCloudflaredBinPath 获取内置 cloudflared 路径
func (s *CloudflareTunnelService) GetCloudflaredBinPath() string {
	return s.manager.GetBinPath()
}


// CreateTunnelRequest 创建隧道请求
type CreateTunnelRequest struct {
	TunnelName string `json:"tunnel_name" binding:"required"` // 隧道名称
	Domain     string `json:"domain" binding:"required"`      // 主域名
	Subdomain  string `json:"subdomain" binding:"required"`   // 子域名前缀
	LocalPort  int    `json:"local_port"`                     // 本地端口，默认 54680
}

// Login 执行登录（会弹出浏览器）
func (s *CloudflareTunnelService) Login() error {
	// 检查 cloudflared 是否安装
	installed, _ := s.CheckCloudflaredInstalled()
	if !installed {
		return errors.New("cloudflared 未安装，请先下载 cloudflared")
	}

	// 如果已登录，直接返回
	if s.CheckLoginStatus() {
		return nil
	}

	// 执行登录命令，会自动弹出浏览器
	cfPath := s.getCloudflaredPath()
	cmd := exec.Command(cfPath, "tunnel", "login")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	// 再次检查登录状态
	if !s.CheckLoginStatus() {
		return errors.New("登录未完成，请在浏览器中完成授权")
	}

	return nil
}

// CreateTunnel 创建隧道（如果未登录会自动弹出浏览器登录）
func (s *CloudflareTunnelService) CreateTunnel(req *CreateTunnelRequest) (*CloudflareTunnelConfig, error) {
	s.tunnelMutex.Lock()
	defer s.tunnelMutex.Unlock()

	// 检查 cloudflared 是否安装
	installed, _ := s.CheckCloudflaredInstalled()
	if !installed {
		return nil, errors.New("cloudflared 未安装，请先下载 cloudflared")
	}

	cfPath := s.getCloudflaredPath()

	// 检查是否已登录，如果未登录则自动执行登录
	if !s.CheckLoginStatus() {
		// 执行登录命令，会自动弹出浏览器
		cmd := exec.Command(cfPath, "tunnel", "login")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("Cloudflare 登录失败: %w", err)
		}

		// 再次检查登录状态
		if !s.CheckLoginStatus() {
			return nil, errors.New("登录未完成，请在浏览器中完成 Cloudflare 授权后重试")
		}
	}

	// 设置默认端口
	if req.LocalPort == 0 {
		req.LocalPort = 54680
	}

	// 构建完整域名
	fullDomain := fmt.Sprintf("%s.%s", req.Subdomain, req.Domain)

	// 检查是否已存在配置
	var existingConfig CloudflareTunnelConfig
	if err := s.db.First(&existingConfig).Error; err == nil {
		// 如果已存在隧道，先删除旧的
		if existingConfig.TunnelID != "" {
			s.deleteTunnelFromCloudflare(existingConfig.TunnelName)
		}
	}

	// 1. 创建隧道
	cmd := exec.Command(cfPath, "tunnel", "create", req.TunnelName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 检查是否是隧道已存在
		if strings.Contains(string(output), "already exists") {
			// 尝试获取已存在的隧道信息
		} else {
			return nil, fmt.Errorf("创建隧道失败: %s", string(output))
		}
	}

	// 2. 获取隧道 ID
	tunnelID, err := s.getTunnelID(req.TunnelName)
	if err != nil {
		return nil, fmt.Errorf("获取隧道ID失败: %w", err)
	}

	// 3. 获取配置文件路径
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".cloudflared")
	configPath := filepath.Join(configDir, fmt.Sprintf("%s-config.yml", req.TunnelName))
	credPath := filepath.Join(configDir, fmt.Sprintf("%s.json", tunnelID))

	// 4. 创建配置文件
	configContent := fmt.Sprintf(`tunnel: %s
credentials-file: %s

ingress:
  - hostname: %s
    service: http://localhost:%d
  - service: http_status:404
`, tunnelID, credPath, fullDomain, req.LocalPort)

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return nil, fmt.Errorf("创建配置文件失败: %w", err)
	}

	// 5. 配置 DNS 路由
	cmd = exec.Command(cfPath, "tunnel", "route", "dns", req.TunnelName, fullDomain)
	output, err = cmd.CombinedOutput()
	if err != nil {
		// DNS 路由可能已存在，忽略错误
		if !strings.Contains(string(output), "already exists") {
			// 记录警告但不失败
			fmt.Printf("DNS 路由配置警告: %s\n", string(output))
		}
	}

	// 6. 保存配置到数据库
	config := CloudflareTunnelConfig{
		TunnelID:   tunnelID,
		TunnelName: req.TunnelName,
		Domain:     req.Domain,
		Subdomain:  req.Subdomain,
		FullDomain: fullDomain,
		LocalPort:  req.LocalPort,
		Status:     "stopped",
		ConfigPath: configPath,
		CredPath:   credPath,
	}

	// 使用 upsert
	if existingConfig.ID > 0 {
		config.ID = existingConfig.ID
		if err := s.db.Save(&config).Error; err != nil {
			return nil, fmt.Errorf("保存配置失败: %w", err)
		}
	} else {
		if err := s.db.Create(&config).Error; err != nil {
			return nil, fmt.Errorf("保存配置失败: %w", err)
		}
	}

	return &config, nil
}

// getTunnelID 获取隧道 ID
func (s *CloudflareTunnelService) getTunnelID(tunnelName string) (string, error) {
	cfPath := s.getCloudflaredPath()
	cmd := exec.Command(cfPath, "tunnel", "list", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var tunnels []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(output, &tunnels); err != nil {
		return "", err
	}

	for _, t := range tunnels {
		if t.Name == tunnelName {
			return t.ID, nil
		}
	}

	return "", errors.New("隧道不存在")
}

// deleteTunnelFromCloudflare 从 Cloudflare 删除隧道
func (s *CloudflareTunnelService) deleteTunnelFromCloudflare(tunnelName string) error {
	cfPath := s.getCloudflaredPath()
	// 先清理连接
	exec.Command(cfPath, "tunnel", "cleanup", tunnelName).Run()
	// 删除隧道
	cmd := exec.Command(cfPath, "tunnel", "delete", tunnelName)
	cmd.Run()
	return nil
}


// StartTunnel 启动隧道
func (s *CloudflareTunnelService) StartTunnel() error {
	s.tunnelMutex.Lock()
	defer s.tunnelMutex.Unlock()

	// 获取配置
	config, err := s.GetConfig()
	if err != nil {
		return err
	}

	if config.TunnelID == "" {
		return errors.New("隧道未配置，请先创建隧道")
	}

	if config.Status == "running" {
		return errors.New("隧道已在运行中")
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(config.ConfigPath); os.IsNotExist(err) {
		return errors.New("配置文件不存在，请重新创建隧道")
	}

	// 创建上下文用于取消
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	// 启动隧道进程
	cfPath := s.getCloudflaredPath()
	cmd := exec.CommandContext(ctx, cfPath, "tunnel", "--config", config.ConfigPath, "run", config.TunnelName)
	
	// 获取输出用于日志
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		s.updateStatus(config.ID, "error", err.Error())
		return fmt.Errorf("启动隧道失败: %w", err)
	}

	s.tunnelCmd = cmd

	// 更新状态
	s.updateStatus(config.ID, "running", "")

	// 异步监控进程
	go func() {
		// 读取输出
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				// 可以记录日志
				fmt.Println("[CF Tunnel]", scanner.Text())
			}
		}()
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				fmt.Println("[CF Tunnel Error]", scanner.Text())
			}
		}()

		// 等待进程结束
		err := cmd.Wait()
		s.tunnelMutex.Lock()
		defer s.tunnelMutex.Unlock()
		
		if err != nil && ctx.Err() == nil {
			// 非正常退出
			s.updateStatus(config.ID, "error", err.Error())
		} else {
			s.updateStatus(config.ID, "stopped", "")
		}
		s.tunnelCmd = nil
	}()

	return nil
}

// StopTunnel 停止隧道
func (s *CloudflareTunnelService) StopTunnel() error {
	s.tunnelMutex.Lock()
	defer s.tunnelMutex.Unlock()

	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}

	if s.tunnelCmd != nil && s.tunnelCmd.Process != nil {
		s.tunnelCmd.Process.Kill()
		s.tunnelCmd = nil
	}

	// 更新状态
	var config CloudflareTunnelConfig
	if err := s.db.First(&config).Error; err == nil {
		s.updateStatus(config.ID, "stopped", "")
	}

	return nil
}

// DeleteTunnel 删除隧道
func (s *CloudflareTunnelService) DeleteTunnel() error {
	// 先停止隧道
	s.StopTunnel()

	s.tunnelMutex.Lock()
	defer s.tunnelMutex.Unlock()

	// 获取配置
	var config CloudflareTunnelConfig
	if err := s.db.First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	// 删除 Cloudflare 隧道
	if config.TunnelName != "" {
		s.deleteTunnelFromCloudflare(config.TunnelName)
	}

	// 删除配置文件
	if config.ConfigPath != "" {
		os.Remove(config.ConfigPath)
	}

	// 删除数据库记录
	return s.db.Delete(&config).Error
}

// updateStatus 更新状态
func (s *CloudflareTunnelService) updateStatus(id uint, status, errorMsg string) {
	s.db.Model(&CloudflareTunnelConfig{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    status,
		"error_msg": errorMsg,
	})
}

// GetStatus 获取隧道状态
type TunnelStatus struct {
	Installed    bool   `json:"installed"`
	Version      string `json:"version"`
	LoggedIn     bool   `json:"logged_in"`
	Configured   bool   `json:"configured"`
	Running      bool   `json:"running"`
	TunnelName   string `json:"tunnel_name"`
	FullDomain   string `json:"full_domain"`
	LocalPort    int    `json:"local_port"`
	ErrorMsg     string `json:"error_msg"`
	NotifyURL    string `json:"notify_url"` // 支付宝回调地址
}

func (s *CloudflareTunnelService) GetStatus() (*TunnelStatus, error) {
	status := &TunnelStatus{}

	// 检查安装
	installed, version := s.CheckCloudflaredInstalled()
	status.Installed = installed
	status.Version = version

	// 检查登录
	status.LoggedIn = s.CheckLoginStatus()

	// 获取配置
	config, err := s.GetConfig()
	if err != nil {
		return status, nil
	}

	status.Configured = config.TunnelID != ""
	status.Running = config.Status == "running"
	status.TunnelName = config.TunnelName
	status.FullDomain = config.FullDomain
	status.LocalPort = config.LocalPort
	status.ErrorMsg = config.ErrorMsg

	// 生成支付宝回调地址
	if config.FullDomain != "" {
		status.NotifyURL = fmt.Sprintf("https://%s/api/v1/payment/alipay/notify", config.FullDomain)
	}

	return status, nil
}

// RestartTunnel 重启隧道
func (s *CloudflareTunnelService) RestartTunnel() error {
	if err := s.StopTunnel(); err != nil {
		return err
	}
	time.Sleep(time.Second) // 等待进程完全停止
	return s.StartTunnel()
}
