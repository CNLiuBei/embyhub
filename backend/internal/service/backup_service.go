// Package service 数据备份服务
package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"feiniu-user-system/internal/config"
)

// BackupService 备份服务
type BackupService struct {
	cfg       *config.Config
	backupDir string
}

// NewBackupService 创建备份服务
func NewBackupService(cfg *config.Config) *BackupService {
	backupDir := "./backups"
	os.MkdirAll(backupDir, 0755)
	return &BackupService{
		cfg:       cfg,
		backupDir: backupDir,
	}
}

// BackupInfo 备份信息
type BackupInfo struct {
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	SizeStr   string `json:"size_str"`
	CreatedAt string `json:"created_at"`
}

// CreateBackup 创建数据库备份
func (s *BackupService) CreateBackup() (*BackupInfo, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s.sql", timestamp)
	filepath := filepath.Join(s.backupDir, filename)

	// 构建 pg_dump 命令
	cmd := exec.Command("pg_dump",
		"-h", s.cfg.Database.Host,
		"-p", fmt.Sprintf("%d", s.cfg.Database.Port),
		"-U", s.cfg.Database.User,
		"-d", s.cfg.Database.DBName,
		"-F", "c", // 自定义格式，支持压缩
		"-f", filepath,
	)

	// 设置密码环境变量
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", s.cfg.Database.Password))

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("备份失败: %s, %v", string(output), err)
	}

	// 获取文件信息
	info, err := os.Stat(filepath)
	if err != nil {
		return nil, fmt.Errorf("获取备份文件信息失败: %v", err)
	}

	return &BackupInfo{
		Filename:  filename,
		Size:      info.Size(),
		SizeStr:   formatFileSize(info.Size()),
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// GetBackupList 获取备份列表
func (s *BackupService) GetBackupList() ([]BackupInfo, error) {
	files, err := os.ReadDir(s.backupDir)
	if err != nil {
		return nil, err
	}

	var backups []BackupInfo
	for _, file := range files {
		if file.IsDir() || !strings.HasPrefix(file.Name(), "backup_") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Filename:  file.Name(),
			Size:      info.Size(),
			SizeStr:   formatFileSize(info.Size()),
			CreatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	// 按时间倒序
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt > backups[j].CreatedAt
	})

	return backups, nil
}

// DeleteBackup 删除备份
func (s *BackupService) DeleteBackup(filename string) error {
	// 安全检查：防止路径遍历
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("无效的文件名")
	}
	if !strings.HasPrefix(filename, "backup_") {
		return fmt.Errorf("无效的备份文件")
	}

	filepath := filepath.Join(s.backupDir, filename)
	return os.Remove(filepath)
}

// GetBackupPath 获取备份文件路径（用于下载）
func (s *BackupService) GetBackupPath(filename string) (string, error) {
	// 安全检查
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return "", fmt.Errorf("无效的文件名")
	}

	path := filepath.Join(s.backupDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("备份文件不存在")
	}

	return path, nil
}

// RestoreBackup 恢复备份（谨慎使用）
func (s *BackupService) RestoreBackup(filename string) error {
	path, err := s.GetBackupPath(filename)
	if err != nil {
		return err
	}

	// 构建 pg_restore 命令
	cmd := exec.Command("pg_restore",
		"-h", s.cfg.Database.Host,
		"-p", fmt.Sprintf("%d", s.cfg.Database.Port),
		"-U", s.cfg.Database.User,
		"-d", s.cfg.Database.DBName,
		"-c", // 先清除再恢复
		path,
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", s.cfg.Database.Password))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("恢复失败: %s, %v", string(output), err)
	}

	return nil
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
