package handler

import (
	"bytes"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

// SetupHandler 安装程序处理器
type SetupHandler struct {
	configPath    string
	validLicenses map[string]bool
}

// NewSetupHandler 创建安装处理器
func NewSetupHandler() *SetupHandler {
	return &SetupHandler{
		configPath: "config/config.yaml",
		validLicenses: map[string]bool{
			"EMBY-HUB-2024-PRO": true,
			"EMBY-FREE-TRIAL":   true,
			"DEMO-LICENSE-KEY":  true,
		},
	}
}

// SetupConfig 安装配置结构（完整配置）
type SetupConfig struct {
	Server   ServerSetupConfig   `yaml:"server" json:"server"`
	Database DatabaseSetupConfig `yaml:"database" json:"database"`
	Redis    RedisSetupConfig    `yaml:"redis" json:"redis"`
	JWT      JWTSetupConfig      `yaml:"jwt" json:"jwt"`
	Emby     EmbySetupConfig     `yaml:"emby" json:"emby"`
	Email    EmailSetupConfig    `yaml:"email" json:"email"`
	Log      LogSetupConfig      `yaml:"log" json:"log"`
	CORS     CORSSetupConfig     `yaml:"cors" json:"cors"`
}

type ServerSetupConfig struct {
	Port int    `yaml:"port" json:"port"`
	Mode string `yaml:"mode" json:"mode"`
}

type DatabaseSetupConfig struct {
	Host            string `yaml:"host" json:"host"`
	Port            int    `yaml:"port" json:"port"`
	User            string `yaml:"user" json:"user"`
	Password        string `yaml:"password" json:"password"`
	DBName          string `yaml:"dbname" json:"dbname"`
	SSLMode         string `yaml:"sslmode" json:"sslmode"`
	MaxIdleConns    int    `yaml:"maxIdleConns" json:"maxIdleConns"`
	MaxOpenConns    int    `yaml:"maxOpenConns" json:"maxOpenConns"`
	ConnMaxLifetime int    `yaml:"connMaxLifetime" json:"connMaxLifetime"`
}

type RedisSetupConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
}

type JWTSetupConfig struct {
	Secret      string `yaml:"secret" json:"secret"`
	ExpireHours int    `yaml:"expireHours" json:"expireHours"`
}

type EmbySetupConfig struct {
	ServerURL string `yaml:"serverUrl" json:"serverUrl"`
	APIKey    string `yaml:"apiKey" json:"apiKey"`
}

type EmailSetupConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	From     string `yaml:"from" json:"from"`
}

type LogSetupConfig struct {
	Level    string `yaml:"level" json:"level"`
	Filename string `yaml:"filename" json:"filename"`
}

type CORSSetupConfig struct {
	AllowOrigins     []string `yaml:"allowOrigins" json:"allowOrigins"`
	AllowMethods     []string `yaml:"allowMethods" json:"allowMethods"`
	AllowHeaders     []string `yaml:"allowHeaders" json:"allowHeaders"`
	ExposeHeaders    []string `yaml:"exposeHeaders" json:"exposeHeaders"`
	AllowCredentials bool     `yaml:"allowCredentials" json:"allowCredentials"`
	MaxAge           int      `yaml:"maxAge" json:"maxAge"`
}

// CheckStatus 检查初始化状态
func (h *SetupHandler) CheckStatus(c *gin.Context) {
	_, err := os.Stat(h.configPath)
	initialized := err == nil

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"initialized": initialized,
		},
	})
}

// GetDefaultConfig 获取默认配置
func (h *SetupHandler) GetDefaultConfig(c *gin.Context) {
	config := SetupConfig{
		Server: ServerSetupConfig{Port: 8080, Mode: "debug"},
		Database: DatabaseSetupConfig{
			Host: "localhost", Port: 5432, User: "postgres", DBName: "embyhub",
			SSLMode: "disable", MaxIdleConns: 10, MaxOpenConns: 100, ConnMaxLifetime: 3600,
		},
		Redis: RedisSetupConfig{Host: "localhost", Port: 6379, Password: "", DB: 0},
		JWT:   JWTSetupConfig{Secret: fmt.Sprintf("embyhub_%d", time.Now().UnixNano()), ExpireHours: 168},
		Emby:  EmbySetupConfig{ServerURL: "http://localhost:8096"},
		Email: EmailSetupConfig{Host: "smtp.example.com", Port: 587},
		Log:   LogSetupConfig{Level: "debug", Filename: "logs/app.log"},
		CORS: CORSSetupConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           86400,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": config,
	})
}

// VerifyLicense 验证授权码
func (h *SetupHandler) VerifyLicense(c *gin.Context) {
	var req struct {
		License string `json:"license"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	license := strings.TrimSpace(strings.ToUpper(req.License))
	if license == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请输入授权码"})
		return
	}

	if h.validLicenses[license] {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "授权验证成功"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 400, "message": "授权码无效，请联系管理员获取"})
}

// TestDatabase 测试数据库连接
func (h *SetupHandler) TestDatabase(c *gin.Context) {
	var req DatabaseSetupConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		req.Host, req.Port, req.User, req.Password)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "连接失败: " + err.Error()})
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "连接失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "数据库连接成功"})
}

// TestEmby 测试 Emby 连接
func (h *SetupHandler) TestEmby(c *gin.Context) {
	var req EmbySetupConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	testURL := fmt.Sprintf("%s/System/Info?api_key=%s", strings.TrimSuffix(req.ServerURL, "/"), req.APIKey)

	resp, err := client.Get(testURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "连接失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": fmt.Sprintf("连接失败: HTTP %d", resp.StatusCode)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Emby 连接成功"})
}

// TestEmail 测试邮件连接
func (h *SetupHandler) TestEmail(c *gin.Context) {
	var req EmailSetupConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	addr := fmt.Sprintf("%s:%d", req.Host, req.Port)
	auth := smtp.PlainAuth("", req.User, req.Password, req.Host)

	var client *smtp.Client
	var err error

	// 465 端口使用 SSL/TLS
	if req.Port == 465 {
		tlsConfig := &tls.Config{ServerName: req.Host}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "SSL连接失败: " + err.Error()})
			return
		}
		client, err = smtp.NewClient(conn, req.Host)
		if err != nil {
			conn.Close()
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "创建客户端失败: " + err.Error()})
			return
		}
	} else {
		client, err = smtp.Dial(addr)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "连接失败: " + err.Error()})
			return
		}
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{ServerName: req.Host}
			if err := client.StartTLS(tlsConfig); err != nil {
				client.Close()
				c.JSON(http.StatusOK, gin.H{"code": 400, "message": "STARTTLS失败: " + err.Error()})
				return
			}
		}
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "认证失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "邮件服务器连接成功"})
}

// FinishSetup 完成安装
func (h *SetupHandler) FinishSetup(c *gin.Context) {
	var req struct {
		Config     SetupConfig `json:"config"`
		AdminUser  string      `json:"admin_user"`
		AdminPass  string      `json:"admin_pass"`
		AdminEmail string      `json:"admin_email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 连接数据库
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		req.Config.Database.Host, req.Config.Database.Port,
		req.Config.Database.User, req.Config.Database.Password)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "数据库连接失败: " + err.Error()})
		return
	}

	// 创建数据库
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", req.Config.Database.DBName))
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "创建数据库失败: " + err.Error()})
		db.Close()
		return
	}
	db.Close()

	// 连接到新数据库
	dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		req.Config.Database.Host, req.Config.Database.Port,
		req.Config.Database.User, req.Config.Database.Password, req.Config.Database.DBName)

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "连接数据库失败: " + err.Error()})
		return
	}
	defer db.Close()

	// 执行初始化 SQL
	if err := h.executeInitSQL(db); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "初始化数据库失败: " + err.Error()})
		return
	}

	// 同步管理员到 Emby 并创建本地账户
	embyUserID, err := h.syncAdminToEmby(req.Config.Emby, req.AdminUser, req.AdminPass)
	if err != nil {
		// Emby 同步失败不阻止安装，只记录警告
		fmt.Printf("警告: 同步管理员到 Emby 失败: %v\n", err)
	}

	// 创建管理员账户（包含 Emby 用户 ID）
	if err := h.createAdminUser(db, req.AdminUser, req.AdminPass, req.AdminEmail, embyUserID); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "创建管理员失败: " + err.Error()})
		return
	}

	// 保存配置文件
	if err := h.saveConfig(req.Config); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "保存配置失败: " + err.Error()})
		return
	}

	// 同步配置到数据库
	if err := h.syncConfigToDatabase(db, req.Config); err != nil {
		fmt.Printf("警告: 同步配置到数据库失败: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "初始化完成"})

	// 延迟重启服务（给响应足够时间返回）
	go func() {
		time.Sleep(2 * time.Second)
		os.Exit(0) // 退出进程，由外部进程管理器重启
	}()
}

func (h *SetupHandler) executeInitSQL(db *sql.DB) error {
	// SQL 文件路径（支持多种运行环境）
	schemaPaths := []string{"database/init_schema.sql", "../database/init_schema.sql"}
	dataPaths := []string{"database/init_data.sql", "../database/init_data.sql"}

	var schemaSQL []byte
	var err error
	for _, path := range schemaPaths {
		schemaSQL, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("读取 schema 文件失败: %v", err)
	}

	var dataSQL []byte
	for _, path := range dataPaths {
		dataSQL, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("读取 data 文件失败: %v", err)
	}

	for _, stmt := range strings.Split(string(schemaSQL), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("执行 SQL 失败: %v", err)
		}
	}

	for _, stmt := range strings.Split(string(dataSQL), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			if !strings.Contains(err.Error(), "duplicate key") {
				return fmt.Errorf("执行 SQL 失败: %v", err)
			}
		}
	}

	return nil
}

func (h *SetupHandler) createAdminUser(db *sql.DB, username, password, email, embyUserID string) error {
	// 使用 bcrypt 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	query := `INSERT INTO users (username, password_hash, email, emby_user_id, role_id, status, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, 1, 1, NOW(), NOW())
			  ON CONFLICT (username) DO UPDATE SET password_hash = $2, email = $3, emby_user_id = $4`
	_, err = db.Exec(query, username, string(hashedPassword), email, embyUserID)
	return err
}

// syncAdminToEmby 同步管理员账户到 Emby
func (h *SetupHandler) syncAdminToEmby(embyConfig EmbySetupConfig, username, password string) (string, error) {
	if embyConfig.ServerURL == "" || embyConfig.APIKey == "" {
		return "", fmt.Errorf("Emby 配置不完整")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// 1. 先检查用户是否已存在
	checkURL := fmt.Sprintf("%s/Users?api_key=%s", strings.TrimSuffix(embyConfig.ServerURL, "/"), embyConfig.APIKey)
	resp, err := client.Get(checkURL)
	if err != nil {
		return "", fmt.Errorf("检查 Emby 用户失败: %v", err)
	}
	defer resp.Body.Close()

	var users []struct {
		ID   string `json:"Id"`
		Name string `json:"Name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return "", fmt.Errorf("解析 Emby 用户列表失败: %v", err)
	}

	// 检查是否已存在同名用户
	for _, u := range users {
		if u.Name == username {
			return u.ID, nil // 已存在，直接返回 ID
		}
	}

	// 2. 创建新用户
	createURL := fmt.Sprintf("%s/Users/New?api_key=%s", strings.TrimSuffix(embyConfig.ServerURL, "/"), embyConfig.APIKey)
	reqBody, _ := json.Marshal(map[string]string{"Name": username})

	resp, err = client.Post(createURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("创建 Emby 用户失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("创建 Emby 用户失败: %s", string(body))
	}

	var newUser struct {
		ID   string `json:"Id"`
		Name string `json:"Name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&newUser); err != nil {
		return "", fmt.Errorf("解析创建结果失败: %v", err)
	}

	// 3. 设置密码
	passURL := fmt.Sprintf("%s/Users/%s/Password?api_key=%s", strings.TrimSuffix(embyConfig.ServerURL, "/"), newUser.ID, embyConfig.APIKey)
	passBody, _ := json.Marshal(map[string]string{"CurrentPw": "", "NewPw": password})

	req, _ := http.NewRequest("POST", passURL, bytes.NewBuffer(passBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		return newUser.ID, fmt.Errorf("设置密码失败: %v", err)
	}
	resp.Body.Close()

	return newUser.ID, nil
}

func (h *SetupHandler) saveConfig(config SetupConfig) error {
	os.MkdirAll("config", 0755)
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(h.configPath, data, 0644)
}

// syncConfigToDatabase 同步配置到数据库
func (h *SetupHandler) syncConfigToDatabase(db *sql.DB, config SetupConfig) error {
	configs := map[string]string{
		"emby_server_url":  config.Emby.ServerURL,
		"emby_api_key":     config.Emby.APIKey,
		"jwt_secret":       config.JWT.Secret,
		"jwt_expire_hours": fmt.Sprintf("%d", config.JWT.ExpireHours),
		"smtp_host":        config.Email.Host,
		"smtp_port":        fmt.Sprintf("%d", config.Email.Port),
		"smtp_user":        config.Email.User,
		"smtp_password":    config.Email.Password,
		"smtp_from":        config.Email.From,
	}

	for key, value := range configs {
		query := `INSERT INTO system_configs (config_key, config_value, updated_at) 
				  VALUES ($1, $2, NOW())
				  ON CONFLICT (config_key) DO UPDATE SET config_value = $2, updated_at = NOW()`
		if _, err := db.Exec(query, key, value); err != nil {
			return fmt.Errorf("更新配置 %s 失败: %v", key, err)
		}
	}
	return nil
}
