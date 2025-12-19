// Package email 邮件发送服务
package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"feiniu-user-system/internal/config"
)

// Config 邮件配置(独立于系统配置)
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// Service 邮件服务
type Service struct {
	cfg             *config.EmailConfig
	templateManager *TemplateManager
	pool            *connectionPool
	logger          Logger
}

// Logger 日志接口
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
}

// defaultLogger 默认日志实现
type defaultLogger struct{}

func (l *defaultLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Printf("[INFO] %s %v", msg, keysAndValues)
}

func (l *defaultLogger) Error(msg string, keysAndValues ...interface{}) {
	log.Printf("[ERROR] %s %v", msg, keysAndValues)
}

func (l *defaultLogger) Warn(msg string, keysAndValues ...interface{}) {
	log.Printf("[WARN] %s %v", msg, keysAndValues)
}

// Message 邮件消息
type Message struct {
	To          []string
	Subject     string
	Body        string
	ContentType string
	Priority    int
}

// connectionPool SMTP持久连接管理
type connectionPool struct {
	cfg       *config.EmailConfig
	mu        sync.Mutex
	client    *smtp.Client
	conn      interface{ Close() error }
	lastUsed  time.Time
	connected bool
	logger    Logger
}

// NewService 创建邮件服务
func NewService(cfg *config.EmailConfig) *Service {
	return NewServiceWithLogger(cfg, &defaultLogger{})
}

// NewServiceWithLogger 创建带日志的邮件服务
func NewServiceWithLogger(cfg *config.EmailConfig, logger Logger) *Service {
	s := &Service{
		cfg:             cfg,
		templateManager: NewTemplateManager(),
		logger:          logger,
	}

	// 初始化连接池
	if cfg.Enabled {
		s.pool = newConnectionPool(cfg, logger)
		s.logger.Info("Email service initialized", "host", cfg.Host, "port", cfg.Port)
	}

	return s
}

// NewServiceWithConfig 使用独立配置创建邮件服务
func NewServiceWithConfig(cfg *Config) *Service {
	emailCfg := &config.EmailConfig{
		Enabled:  true,
		Host:     cfg.Host,
		Port:     cfg.Port,
		Username: cfg.Username,
		Password: cfg.Password,
		From:     cfg.From,
		FromName: cfg.FromName,
	}
	return NewService(emailCfg)
}

// newConnectionPool 创建连接池
func newConnectionPool(cfg *config.EmailConfig, logger Logger) *connectionPool {
	pool := &connectionPool{
		cfg:    cfg,
		logger: logger,
	}
	// 尝试建立初始连接
	go pool.connect()
	return pool
}

// connect 建立SMTP连接
func (p *connectionPool) connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 如果已连接，先关闭
	if p.client != nil {
		p.client.Close()
		p.client = nil
	}
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
	p.connected = false

	addr := fmt.Sprintf("%s:%d", p.cfg.Host, p.cfg.Port)
	auth := smtp.PlainAuth("", p.cfg.Username, p.cfg.Password, p.cfg.Host)

	var client *smtp.Client
	var conn interface{ Close() error }

	if p.cfg.Port == 465 {
		// TLS连接
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         p.cfg.Host,
		}
		tlsConn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			p.logger.Error("Failed to connect to SMTP server", "error", err)
			return fmt.Errorf("TLS连接失败: %v", err)
		}
		conn = tlsConn

		client, err = smtp.NewClient(tlsConn, p.cfg.Host)
		if err != nil {
			tlsConn.Close()
			p.logger.Error("Failed to create SMTP client", "error", err)
			return fmt.Errorf("创建SMTP客户端失败: %v", err)
		}
	} else {
		// 普通连接
		var err error
		client, err = smtp.Dial(addr)
		if err != nil {
			p.logger.Error("Failed to connect to SMTP server", "error", err)
			return fmt.Errorf("SMTP连接失败: %v", err)
		}
	}

	// 认证
	if err := client.Auth(auth); err != nil {
		client.Close()
		if conn != nil {
			conn.Close()
		}
		p.logger.Error("SMTP authentication failed", "error", err)
		return fmt.Errorf("SMTP认证失败: %v", err)
	}

	p.client = client
	p.conn = conn
	p.connected = true
	p.lastUsed = time.Now()
	p.logger.Info("SMTP connection established", "host", p.cfg.Host)

	return nil
}

// getClient 获取SMTP客户端，如果断开则重连
func (p *connectionPool) getClient() (*smtp.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查连接是否有效
	if p.connected && p.client != nil {
		// 使用NOOP命令检测连接是否有效
		if err := p.client.Noop(); err == nil {
			p.lastUsed = time.Now()
			return p.client, nil
		}
		// 连接已断开
		p.logger.Warn("SMTP connection lost, reconnecting...")
		p.connected = false
	}

	// 需要重新连接，先解锁再加锁
	p.mu.Unlock()
	err := p.connect()
	p.mu.Lock()

	if err != nil {
		return nil, err
	}

	return p.client, nil
}

// resetConnection 重置连接（发送完邮件后）
func (p *connectionPool) resetConnection() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		// 重置SMTP会话以便发送下一封邮件
		p.client.Reset()
	}
}

// ClosePool 关闭连接池
func (p *connectionPool) ClosePool() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		p.client.Quit()
		p.client = nil
	}
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
	p.connected = false
}

// Send 发送邮件消息
func (s *Service) Send(msg *Message) error {
	if !s.cfg.Enabled {
		s.logger.Warn("Email service is disabled")
		return ErrEmailNotEnabled
	}

	if len(msg.To) == 0 {
		return ErrInvalidRecipient
	}

	// 构建邮件头
	from := s.cfg.From
	if s.cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.From)
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(msg.To, ", ")
	headers["Subject"] = msg.Subject
	headers["MIME-Version"] = "1.0"
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	if msg.ContentType == "" {
		msg.ContentType = "text/html; charset=UTF-8"
	}
	headers["Content-Type"] = msg.ContentType

	if msg.Priority > 0 {
		headers["X-Priority"] = fmt.Sprintf("%d", msg.Priority)
	}

	// 构建邮件内容
	var msgBuilder strings.Builder
	for k, v := range headers {
		msgBuilder.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(msg.Body)

	// 发送邮件，带重试
	return s.sendWithRetry(msg.To, []byte(msgBuilder.String()), 3)
}

// sendWithRetry 带重试的发送
func (s *Service) sendWithRetry(to []string, content []byte, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// 重试前等待
			waitTime := time.Duration(i) * time.Second
			s.logger.Info("Retrying email send", "attempt", i+1, "wait", waitTime)
			time.Sleep(waitTime)
		}

		err := s.sendMail(to, content)
		if err == nil {
			if i > 0 {
				s.logger.Info("Email sent successfully after retry", "attempts", i+1)
			}
			return nil
		}

		lastErr = err
		s.logger.Error("Failed to send email", "attempt", i+1, "error", err)
	}

	return fmt.Errorf("%w: %v", ErrSendFailed, lastErr)
}

// sendMail 发送邮件（使用持久连接）
func (s *Service) sendMail(to []string, content []byte) error {
	// 尝试使用连接池的持久连接
	if s.pool != nil {
		client, err := s.pool.getClient()
		if err == nil && client != nil {
			// 使用持久连接发送
			if err := s.sendWithClient(client, to, content); err == nil {
				s.pool.resetConnection()
				return nil
			}
			// 发送失败，标记连接需要重建
			s.logger.Warn("Persistent connection send failed, falling back to new connection")
		}
	}

	// 回退到创建新连接发送
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	// 根据端口选择发送方式
	if s.cfg.Port == 465 {
		return s.sendWithTLS(addr, auth, to, content)
	}

	// 使用标准 SMTP
	return smtp.SendMail(addr, auth, s.cfg.From, to, content)
}

// sendWithClient 使用已有客户端发送邮件
func (s *Service) sendWithClient(client *smtp.Client, to []string, content []byte) error {
	if err := client.Mail(s.cfg.From); err != nil {
		return err
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(content)
	if err != nil {
		return err
	}

	return w.Close()
}

// sendWithTLS 使用TLS发送邮件
func (s *Service) sendWithTLS(addr string, auth smtp.Auth, to []string, content []byte) error {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.cfg.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("%w: %v", ErrAuthFailed, err)
	}

	if err := client.Mail(s.cfg.From); err != nil {
		return err
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(content)
	if err != nil {
		return err
	}

	if err = w.Close(); err != nil {
		return err
	}

	return client.Quit()
}

// SendHTML 发送HTML邮件
func (s *Service) SendHTML(to, subject, body string) error {
	return s.Send(&Message{
		To:          []string{to},
		Subject:     subject,
		Body:        body,
		ContentType: "text/html; charset=UTF-8",
	})
}

// SendTemplate 使用模板发送邮件
func (s *Service) SendTemplate(to, subject string, templateType TemplateType, data interface{}) error {
	body, err := s.templateManager.Render(templateType, data)
	if err != nil {
		s.logger.Error("Failed to render template", "template", templateType, "error", err)
		return err
	}

	return s.SendHTML(to, subject, body)
}

// SendTestEmail 发送测试邮件
func (s *Service) SendTestEmail(to string) error {
	if !s.cfg.Enabled {
		return ErrEmailNotEnabled
	}

	subject := "邮件服务测试"
	data := map[string]interface{}{
		"FromName": s.cfg.FromName,
	}

	body, err := s.templateManager.Render(TemplateTest, data)
	if err != nil {
		// 如果模板失败，使用简单HTML
		body = `
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px;">
    <div style="max-width: 500px; margin: 0 auto; background: #f5f5f5; border-radius: 10px; padding: 30px;">
        <h2 style="color: #333; text-align: center;">✅ 邮件服务测试成功</h2>
        <p style="color: #666; text-align: center;">如果您收到这封邮件，说明邮件服务配置正确。</p>
    </div>
</body>
</html>`
	}

	return s.SendHTML(to, subject, body)
}

// SendRegisterCode 发送注册验证码
func (s *Service) SendRegisterCode(to, code string) error {
	if !s.cfg.Enabled {
		s.logger.Error("Email service is not enabled")
		return ErrEmailNotEnabled
	}

	s.logger.Info("Sending register code", "to", to, "code", code)

	subject := "注册验证码"
	data := map[string]interface{}{
		"FromName": s.cfg.FromName,
		"Code":     code,
	}

	err := s.SendTemplate(to, subject, TemplateRegisterCode, data)
	if err != nil {
		s.logger.Error("Failed to send register code", "to", to, "error", err)
	} else {
		s.logger.Info("Register code sent successfully", "to", to)
	}
	return err
}

// SendResetPasswordCode 发送密码重置验证码
func (s *Service) SendResetPasswordCode(to, code string) error {
	if !s.cfg.Enabled {
		return ErrEmailNotEnabled
	}

	subject := "密码重置验证码"
	data := map[string]interface{}{
		"FromName": s.cfg.FromName,
		"Code":     code,
	}

	return s.SendTemplate(to, subject, TemplateResetPasswordCode, data)
}

// SendWelcome 发送注册成功欢迎邮件
func (s *Service) SendWelcome(to, username string) error {
	if !s.cfg.Enabled {
		return ErrEmailNotEnabled
	}

	subject := fmt.Sprintf("欢迎加入 %s", s.cfg.FromName)
	data := map[string]interface{}{
		"FromName": s.cfg.FromName,
		"Username": username,
	}

	return s.SendTemplate(to, subject, TemplateWelcome, data)
}

// SendPasswordChanged 发送密码修改成功通知
func (s *Service) SendPasswordChanged(to, username string) error {
	if !s.cfg.Enabled {
		return ErrEmailNotEnabled
	}

	subject := "密码修改成功通知"
	data := map[string]interface{}{
		"FromName": s.cfg.FromName,
		"Username": username,
	}

	return s.SendTemplate(to, subject, TemplatePasswordChanged, data)
}

// SendMembershipExpireSoon 发送会员即将到期提醒
func (s *Service) SendMembershipExpireSoon(to, username string, daysLeft int) error {
	if !s.cfg.Enabled {
		return ErrEmailNotEnabled
	}

	subject := fmt.Sprintf("会员即将到期提醒 - 还剩%d天", daysLeft)
	data := map[string]interface{}{
		"FromName": s.cfg.FromName,
		"Username": username,
		"DaysLeft": daysLeft,
	}

	return s.SendTemplate(to, subject, TemplateMembershipExpire, data)
}

// SendLoginAlert 发送异地登录提醒
func (s *Service) SendLoginAlert(to, username, ip, location, device string) error {
	if !s.cfg.Enabled {
		return ErrEmailNotEnabled
	}

	subject := "账户登录提醒"
	data := map[string]interface{}{
		"FromName": s.cfg.FromName,
		"Username": username,
		"IP":       ip,
		"Location": location,
		"Device":   device,
	}

	return s.SendTemplate(to, subject, TemplateLoginAlert, data)
}

// SendMemberActivated 发送会员开通/续费成功通知
func (s *Service) SendMemberActivated(to, username string, days int, expireDate string, isRenew bool) error {
	if !s.cfg.Enabled {
		return ErrEmailNotEnabled
	}

	action := "开通"
	if isRenew {
		action = "续费"
	}
	subject := fmt.Sprintf("会员%s成功通知", action)

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px; background: #f5f5f5;">
    <div style="max-width: 500px; margin: 0 auto; background: white; border-radius: 10px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
        <h2 style="color: #52c41a; text-align: center;">🎉 会员%s成功</h2>
        <p style="color: #333;">尊敬的 <strong>%s</strong>，您好！</p>
        <p style="color: #666;">恭喜您成功%s会员，详情如下：</p>
        <div style="background: #f0f9ff; border-radius: 8px; padding: 15px; margin: 20px 0;">
            <p style="margin: 5px 0;"><strong>会员时长：</strong>%d 天</p>
            <p style="margin: 5px 0;"><strong>到期时间：</strong>%s</p>
        </div>
        <p style="color: #666;">感谢您的支持，祝您观影愉快！</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px; text-align: center;">%s</p>
    </div>
</body>
</html>`, action, username, action, days, expireDate, s.cfg.FromName)

	return s.SendHTML(to, subject, body)
}

// SendAccountDisabled 发送账户被禁用通知
func (s *Service) SendAccountDisabled(to, username, reason string) error {
	if !s.cfg.Enabled {
		return ErrEmailNotEnabled
	}

	subject := "账户已禁用通知"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px; background: #f5f5f5;">
    <div style="max-width: 500px; margin: 0 auto; background: white; border-radius: 10px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
        <h2 style="color: #ff4d4f; text-align: center;">⚠️ 账户已禁用</h2>
        <p style="color: #333;">尊敬的 <strong>%s</strong>，您好！</p>
        <p style="color: #666;">您的账户已被禁用，原因如下：</p>
        <div style="background: #fff2f0; border: 1px solid #ffccc7; border-radius: 8px; padding: 15px; margin: 20px 0;">
            <p style="color: #ff4d4f; margin: 0;">%s</p>
        </div>
        <p style="color: #666;">如需恢复账户，请使用卡密续费或联系管理员。</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px; text-align: center;">%s</p>
    </div>
</body>
</html>`, username, reason, s.cfg.FromName)

	return s.SendHTML(to, subject, body)
}

// SendAccountEnabled 发送账户已启用通知
func (s *Service) SendAccountEnabled(to, username string) error {
	if !s.cfg.Enabled {
		return ErrEmailNotEnabled
	}

	subject := "账户已恢复通知"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px; background: #f5f5f5;">
    <div style="max-width: 500px; margin: 0 auto; background: white; border-radius: 10px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
        <h2 style="color: #52c41a; text-align: center;">✅ 账户已恢复</h2>
        <p style="color: #333;">尊敬的 <strong>%s</strong>，您好！</p>
        <p style="color: #666;">您的账户已恢复正常，现在可以正常登录使用了。</p>
        <p style="color: #666;">感谢您的支持！</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px; text-align: center;">%s</p>
    </div>
</body>
</html>`, username, s.cfg.FromName)

	return s.SendHTML(to, subject, body)
}

// SendPasswordReset 发送密码被重置通知（管理员重置）
func (s *Service) SendPasswordReset(to, username, newPassword string) error {
	if !s.cfg.Enabled {
		return ErrEmailNotEnabled
	}

	subject := "密码重置通知"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px; background: #f5f5f5;">
    <div style="max-width: 500px; margin: 0 auto; background: white; border-radius: 10px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
        <h2 style="color: #1890ff; text-align: center;">🔐 密码已重置</h2>
        <p style="color: #333;">尊敬的 <strong>%s</strong>，您好！</p>
        <p style="color: #666;">管理员已为您重置密码，新密码如下：</p>
        <div style="background: #e6f7ff; border: 1px solid #91d5ff; border-radius: 8px; padding: 15px; margin: 20px 0; text-align: center;">
            <p style="color: #1890ff; font-size: 18px; font-weight: bold; margin: 0; font-family: monospace;">%s</p>
        </div>
        <p style="color: #ff4d4f; font-size: 12px;">⚠️ 请登录后立即修改密码以确保账户安全</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px; text-align: center;">%s</p>
    </div>
</body>
</html>`, username, newPassword, s.cfg.FromName)

	return s.SendHTML(to, subject, body)
}

// SendMemberExpired 发送会员已过期通知
func (s *Service) SendMemberExpired(to, username string) error {
	if !s.cfg.Enabled {
		return ErrEmailNotEnabled
	}

	subject := "会员已到期通知"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px; background: #f5f5f5;">
    <div style="max-width: 500px; margin: 0 auto; background: white; border-radius: 10px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
        <h2 style="color: #faad14; text-align: center;">⏰ 会员已到期</h2>
        <p style="color: #333;">尊敬的 <strong>%s</strong>，您好！</p>
        <p style="color: #666;">您的会员已到期，账户已被暂时禁用。</p>
        <p style="color: #666;">如需继续使用，请使用卡密续费后重新登录。</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px; text-align: center;">%s</p>
    </div>
</body>
</html>`, username, s.cfg.FromName)

	return s.SendHTML(to, subject, body)
}

// SendChangeEmailCode 发送修改邮箱验证码
func (s *Service) SendChangeEmailCode(to, code string) error {
	if !s.cfg.Enabled {
		return ErrEmailNotEnabled
	}

	subject := "邮箱修改验证码"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px; background: #f5f5f5;">
    <div style="max-width: 500px; margin: 0 auto; background: white; border-radius: 10px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
        <h2 style="color: #1890ff; text-align: center;">📧 邮箱修改验证码</h2>
        <p style="color: #333;">您好！</p>
        <p style="color: #666;">您正在修改账户绑定邮箱，验证码如下：</p>
        <div style="background: #e6f7ff; border: 1px solid #91d5ff; border-radius: 8px; padding: 20px; margin: 20px 0; text-align: center;">
            <p style="color: #1890ff; font-size: 32px; font-weight: bold; margin: 0; letter-spacing: 8px;">%s</p>
        </div>
        <p style="color: #666;">验证码有效期为 <strong>10分钟</strong>，请尽快完成验证。</p>
        <p style="color: #ff4d4f; font-size: 12px;">⚠️ 如果这不是您本人的操作，请忽略此邮件。</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px; text-align: center;">%s</p>
    </div>
</body>
</html>`, code, s.cfg.FromName)

	return s.SendHTML(to, subject, body)
}

// SendEmailChanged 发送邮箱变更通知（发送到旧邮箱）
func (s *Service) SendEmailChanged(to, username, newEmail string) error {
	if !s.cfg.Enabled {
		return ErrEmailNotEnabled
	}

	subject := "邮箱变更通知"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px; background: #f5f5f5;">
    <div style="max-width: 500px; margin: 0 auto; background: white; border-radius: 10px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
        <h2 style="color: #faad14; text-align: center;">📧 邮箱已变更</h2>
        <p style="color: #333;">尊敬的 <strong>%s</strong>，您好！</p>
        <p style="color: #666;">您的账户绑定邮箱已变更为：</p>
        <div style="background: #fffbe6; border: 1px solid #ffe58f; border-radius: 8px; padding: 15px; margin: 20px 0; text-align: center;">
            <p style="color: #d48806; font-size: 16px; font-weight: bold; margin: 0;">%s</p>
        </div>
        <p style="color: #ff4d4f; font-size: 12px;">⚠️ 如果这不是您本人的操作，请立即联系管理员。</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
        <p style="color: #999; font-size: 12px; text-align: center;">%s</p>
    </div>
</body>
</html>`, username, newEmail, s.cfg.FromName)

	return s.SendHTML(to, subject, body)
}

// Close 关闭服务
func (s *Service) Close() error {
	if s.pool != nil {
		s.pool.ClosePool()
	}
	s.logger.Info("Email service closed")
	return nil
}

// IsEnabled 检查服务是否启用
func (s *Service) IsEnabled() bool {
	return s.cfg.Enabled
}

// GetTemplateManager 获取模板管理器
func (s *Service) GetTemplateManager() *TemplateManager {
	return s.templateManager
}
