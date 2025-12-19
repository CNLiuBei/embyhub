// Package config 配置管理模块
// 负责加载和管理应用程序的所有配置项
package config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config 应用程序主配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Log      LogConfig      `yaml:"log"`
	Security SecurityConfig `yaml:"security"`
	Upload   UploadConfig   `yaml:"upload"`
	Member   MemberConfig   `yaml:"member"`
	Emby     EmbyConfig     `yaml:"emby"`
	Email    EmailConfig    `yaml:"email"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	FromName string `yaml:"from_name"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `yaml:"port"`
	Mode         string `yaml:"mode"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

// DatabaseConfig PostgreSQL 数据库配置
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	DBName          string `yaml:"dbname"`
	SSLMode         string `yaml:"sslmode"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

// RedisConfig Redis 缓存配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// JWTConfig JWT 认证配置
type JWTConfig struct {
	Secret        string `yaml:"secret"`
	AccessExpire  int    `yaml:"access_expire"`
	RefreshExpire int    `yaml:"refresh_expire"`
	Issuer        string `yaml:"issuer"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	AESKey            string    `yaml:"aes_key"`
	RateLimit         RateLimit `yaml:"rate_limit"`
	PasswordMinLength int       `yaml:"password_min_length"`
}

// RateLimit 限流配置
type RateLimit struct {
	Login int `yaml:"login"`
	API   int `yaml:"api"`
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	AvatarPath   string   `yaml:"avatar_path"`
	MaxSize      int64    `yaml:"max_size"`
	AllowedTypes []string `yaml:"allowed_types"`
}

// MemberConfig 会员配置
type MemberConfig struct {
	Levels []MemberLevel `yaml:"levels"`
}

// MemberLevel 会员等级
type MemberLevel struct {
	ID         int    `yaml:"id"`
	Name       string `yaml:"name"`
	WatchLimit int    `yaml:"watch_limit"`
	AdFree     bool   `yaml:"ad_free"`
	Quality4K  bool   `yaml:"quality_4k"`
}

// EmbyConfig Emby API配置
type EmbyConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Mode      string `yaml:"mode"`       // emby 或 feiniu
	BaseURL   string `yaml:"base_url"`
	APIKey    string `yaml:"api_key"`    // Emby API密钥
	AdminUser string `yaml:"admin_user"` // 飞牛模式使用
	AdminPass string `yaml:"admin_pass"` // 飞牛模式使用
}

// IsEmbyMode 是否为Emby官方模式
func (c *EmbyConfig) IsEmbyMode() bool {
	return c.Mode == "emby" || c.Mode == ""
}

// IsFeiniuMode 是否为飞牛影视模式
func (c *EmbyConfig) IsFeiniuMode() bool {
	return c.Mode == "feiniu"
}

var (
	cfg  *Config
	once sync.Once
)

// Load 加载配置文件
func Load(path string) (*Config, error) {
	var err error
	once.Do(func() {
		cfg = &Config{}
		var data []byte
		data, err = os.ReadFile(path)
		if err != nil {
			return
		}
		err = yaml.Unmarshal(data, cfg)
	})
	return cfg, err
}

// Get 获取配置实例
func Get() *Config {
	return cfg
}
