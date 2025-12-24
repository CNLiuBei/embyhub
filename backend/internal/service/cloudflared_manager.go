// Package service Cloudflared 二进制管理器
// 自动下载和管理 cloudflared 二进制文件，无需用户手动安装
package service

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// cloudflared 下载地址模板
	cloudflaredDownloadURL = "https://github.com/cloudflare/cloudflared/releases/latest/download/%s"
)

// CloudflaredManager 管理 cloudflared 二进制文件
type CloudflaredManager struct {
	binDir  string // 二进制文件存放目录
	binPath string // cloudflared 二进制文件路径
}

// NewCloudflaredManager 创建管理器
func NewCloudflaredManager() *CloudflaredManager {
	// 获取项目根目录下的 bin 目录
	execPath, _ := os.Executable()
	projectRoot := filepath.Dir(filepath.Dir(execPath))
	
	// 如果是开发环境，使用当前工作目录
	if strings.Contains(execPath, "go-build") {
		wd, _ := os.Getwd()
		projectRoot = wd
	}
	
	binDir := filepath.Join(projectRoot, "bin")
	
	// 根据操作系统确定二进制文件名
	binName := "cloudflared"
	if runtime.GOOS == "windows" {
		binName = "cloudflared.exe"
	}
	
	return &CloudflaredManager{
		binDir:  binDir,
		binPath: filepath.Join(binDir, binName),
	}
}

// GetBinPath 获取 cloudflared 二进制文件路径
func (m *CloudflaredManager) GetBinPath() string {
	return m.binPath
}

// IsInstalled 检查是否已安装
func (m *CloudflaredManager) IsInstalled() bool {
	_, err := os.Stat(m.binPath)
	return err == nil
}

// GetVersion 获取版本
func (m *CloudflaredManager) GetVersion() (string, error) {
	if !m.IsInstalled() {
		return "", fmt.Errorf("cloudflared 未安装")
	}
	
	cmd := exec.Command(m.binPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getDownloadFileName 根据操作系统和架构获取下载文件名
func (m *CloudflaredManager) getDownloadFileName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH
	
	switch os {
	case "darwin":
		if arch == "arm64" {
			return "cloudflared-darwin-arm64.tgz"
		}
		return "cloudflared-darwin-amd64.tgz"
	case "linux":
		switch arch {
		case "arm64":
			return "cloudflared-linux-arm64"
		case "arm":
			return "cloudflared-linux-arm"
		case "386":
			return "cloudflared-linux-386"
		default:
			return "cloudflared-linux-amd64"
		}
	case "windows":
		if arch == "386" {
			return "cloudflared-windows-386.exe"
		}
		return "cloudflared-windows-amd64.exe"
	default:
		return "cloudflared-linux-amd64"
	}
}


// Download 下载 cloudflared
func (m *CloudflaredManager) Download(progressCallback func(downloaded, total int64)) error {
	// 创建 bin 目录
	if err := os.MkdirAll(m.binDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	
	fileName := m.getDownloadFileName()
	downloadURL := fmt.Sprintf(cloudflaredDownloadURL, fileName)
	
	// 下载文件
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}
	
	// 根据文件类型处理
	if strings.HasSuffix(fileName, ".tgz") {
		// macOS: 解压 tgz 文件
		return m.extractTgz(resp.Body, progressCallback, resp.ContentLength)
	} else if strings.HasSuffix(fileName, ".zip") {
		// 解压 zip 文件
		return m.extractZip(resp.Body, progressCallback, resp.ContentLength)
	} else {
		// Linux/Windows: 直接保存二进制文件
		return m.saveBinary(resp.Body, progressCallback, resp.ContentLength)
	}
}

// saveBinary 保存二进制文件
func (m *CloudflaredManager) saveBinary(reader io.Reader, progressCallback func(downloaded, total int64), total int64) error {
	outFile, err := os.Create(m.binPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer outFile.Close()
	
	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			_, writeErr := outFile.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("写入文件失败: %w", writeErr)
			}
			downloaded += int64(n)
			if progressCallback != nil {
				progressCallback(downloaded, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取数据失败: %w", err)
		}
	}
	
	// 设置可执行权限
	if runtime.GOOS != "windows" {
		if err := os.Chmod(m.binPath, 0755); err != nil {
			return fmt.Errorf("设置权限失败: %w", err)
		}
	}
	
	return nil
}

// extractTgz 解压 tgz 文件 (macOS)
func (m *CloudflaredManager) extractTgz(reader io.Reader, progressCallback func(downloaded, total int64), total int64) error {
	// 先下载到临时文件
	tmpFile, err := os.CreateTemp("", "cloudflared-*.tgz")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	
	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			_, writeErr := tmpFile.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("写入临时文件失败: %w", writeErr)
			}
			downloaded += int64(n)
			if progressCallback != nil {
				progressCallback(downloaded, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取数据失败: %w", err)
		}
	}
	
	// 重新打开文件进行解压
	tmpFile.Seek(0, 0)
	
	gzReader, err := gzip.NewReader(tmpFile)
	if err != nil {
		return fmt.Errorf("解压 gzip 失败: %w", err)
	}
	defer gzReader.Close()
	
	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取 tar 失败: %w", err)
		}
		
		// 只提取 cloudflared 二进制文件
		if header.Typeflag == tar.TypeReg && strings.Contains(header.Name, "cloudflared") {
			outFile, err := os.Create(m.binPath)
			if err != nil {
				return fmt.Errorf("创建文件失败: %w", err)
			}
			
			_, err = io.Copy(outFile, tarReader)
			outFile.Close()
			if err != nil {
				return fmt.Errorf("解压文件失败: %w", err)
			}
			
			// 设置可执行权限
			if err := os.Chmod(m.binPath, 0755); err != nil {
				return fmt.Errorf("设置权限失败: %w", err)
			}
			
			return nil
		}
	}
	
	return fmt.Errorf("在压缩包中未找到 cloudflared")
}

// extractZip 解压 zip 文件
func (m *CloudflaredManager) extractZip(reader io.Reader, progressCallback func(downloaded, total int64), total int64) error {
	// 先下载到临时文件
	tmpFile, err := os.CreateTemp("", "cloudflared-*.zip")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	
	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			_, writeErr := tmpFile.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("写入临时文件失败: %w", writeErr)
			}
			downloaded += int64(n)
			if progressCallback != nil {
				progressCallback(downloaded, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取数据失败: %w", err)
		}
	}
	
	// 获取文件大小
	fileInfo, _ := tmpFile.Stat()
	
	// 打开 zip 文件
	zipReader, err := zip.NewReader(tmpFile, fileInfo.Size())
	if err != nil {
		return fmt.Errorf("打开 zip 失败: %w", err)
	}
	
	for _, file := range zipReader.File {
		if strings.Contains(file.Name, "cloudflared") {
			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("打开压缩文件失败: %w", err)
			}
			
			outFile, err := os.Create(m.binPath)
			if err != nil {
				rc.Close()
				return fmt.Errorf("创建文件失败: %w", err)
			}
			
			_, err = io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
			
			if err != nil {
				return fmt.Errorf("解压文件失败: %w", err)
			}
			
			// 设置可执行权限
			if runtime.GOOS != "windows" {
				if err := os.Chmod(m.binPath, 0755); err != nil {
					return fmt.Errorf("设置权限失败: %w", err)
				}
			}
			
			return nil
		}
	}
	
	return fmt.Errorf("在压缩包中未找到 cloudflared")
}

// Remove 删除 cloudflared
func (m *CloudflaredManager) Remove() error {
	if m.IsInstalled() {
		return os.Remove(m.binPath)
	}
	return nil
}

// GetConfigDir 获取 cloudflared 配置目录
func (m *CloudflaredManager) GetConfigDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".cloudflared")
}

// IsLoggedIn 检查是否已登录
func (m *CloudflaredManager) IsLoggedIn() bool {
	certPath := filepath.Join(m.GetConfigDir(), "cert.pem")
	_, err := os.Stat(certPath)
	return err == nil
}
