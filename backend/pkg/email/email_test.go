package email

import (
	"testing"
	"time"

	"feiniu-user-system/internal/config"
)

// mockLogger 测试用日志
type mockLogger struct {
	logs []string
}

func (m *mockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.logs = append(m.logs, msg)
}

func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.logs = append(m.logs, msg)
}

func (m *mockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.logs = append(m.logs, msg)
}

func TestNewService(t *testing.T) {
	cfg := &config.EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     465,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		FromName: "Test Service",
	}

	service := NewService(cfg)
	if service == nil {
		t.Fatal("NewService returned nil")
	}

	if !service.IsEnabled() {
		t.Error("Service should be enabled")
	}

	if service.GetTemplateManager() == nil {
		t.Error("Template manager should not be nil")
	}
}

func TestNewServiceWithConfig(t *testing.T) {
	cfg := &Config{
		Host:     "smtp.example.com",
		Port:     465,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		FromName: "Test Service",
	}

	service := NewServiceWithConfig(cfg)
	if service == nil {
		t.Fatal("NewServiceWithConfig returned nil")
	}

	if !service.IsEnabled() {
		t.Error("Service should be enabled")
	}
}

func TestServiceDisabled(t *testing.T) {
	cfg := &config.EmailConfig{
		Enabled: false,
	}

	service := NewService(cfg)
	err := service.SendTestEmail("test@example.com")

	if err != ErrEmailNotEnabled {
		t.Errorf("Expected ErrEmailNotEnabled, got %v", err)
	}
}

func TestTemplateManager(t *testing.T) {
	tm := NewTemplateManager()

	// 测试所有默认模板是否加载
	templates := []TemplateType{
		TemplateTest,
		TemplateRegisterCode,
		TemplateResetPasswordCode,
		TemplateWelcome,
		TemplatePasswordChanged,
		TemplateMembershipExpire,
		TemplateLoginAlert,
	}

	for _, tmplType := range templates {
		if !tm.HasTemplate(tmplType) {
			t.Errorf("Template %s should exist", tmplType)
		}
	}
}

func TestTemplateRender(t *testing.T) {
	tm := NewTemplateManager()

	// 测试注册验证码模板
	data := map[string]interface{}{
		"FromName": "Test Service",
		"Code":     "123456",
	}

	html, err := tm.Render(TemplateRegisterCode, data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if html == "" {
		t.Error("Rendered HTML should not be empty")
	}

	// 检查是否包含验证码
	if !contains(html, "123456") {
		t.Error("Rendered HTML should contain the code")
	}
}

func TestTemplateRenderWelcome(t *testing.T) {
	tm := NewTemplateManager()

	data := map[string]interface{}{
		"FromName": "Test Service",
		"Username": "testuser",
	}

	html, err := tm.Render(TemplateWelcome, data)
	if err != nil {
		t.Fatalf("Failed to render welcome template: %v", err)
	}

	if !contains(html, "testuser") {
		t.Error("Rendered HTML should contain username")
	}
}

func TestTemplateNotFound(t *testing.T) {
	tm := NewTemplateManager()

	_, err := tm.Render("nonexistent", nil)
	if err != ErrTemplateNotFound {
		t.Errorf("Expected ErrTemplateNotFound, got %v", err)
	}
}

func TestMessage(t *testing.T) {
	msg := &Message{
		To:          []string{"test@example.com"},
		Subject:     "Test Subject",
		Body:        "<h1>Test Body</h1>",
		ContentType: "text/html; charset=UTF-8",
		Priority:    1,
	}

	if len(msg.To) != 1 {
		t.Error("Message should have one recipient")
	}

	if msg.Subject != "Test Subject" {
		t.Error("Message subject mismatch")
	}
}

func TestQueueCreation(t *testing.T) {
	cfg := &config.EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     465,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		FromName: "Test Service",
	}

	logger := &mockLogger{}
	service := NewServiceWithLogger(cfg, logger)
	queue := NewQueue(service, 2, 10)

	if queue == nil {
		t.Fatal("NewQueue returned nil")
	}

	metrics := queue.GetMetrics()
	if metrics.TotalEnqueued != 0 {
		t.Error("Initial queue should be empty")
	}

	// 清理
	queue.Stop(time.Second)
	service.Close()
}

func TestQueueMetrics(t *testing.T) {
	cfg := &config.EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     465,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		FromName: "Test Service",
	}

	service := NewService(cfg)
	queue := NewQueue(service, 2, 10)

	// 检查初始指标
	metrics := queue.GetMetrics()
	if metrics.TotalEnqueued != 0 {
		t.Error("TotalEnqueued should be 0")
	}

	if metrics.TotalSent != 0 {
		t.Error("TotalSent should be 0")
	}

	// 清理
	queue.Stop(time.Second)
	service.Close()
}

func TestQueueHealthy(t *testing.T) {
	cfg := &config.EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     465,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		FromName: "Test Service",
	}

	service := NewService(cfg)
	queue := NewQueue(service, 2, 10)

	// 新队列应该是健康的
	if !queue.IsHealthy() {
		t.Error("New queue should be healthy")
	}

	// 清理
	queue.Stop(time.Second)
	service.Close()
}

func TestConnectionPoolCreation(t *testing.T) {
	cfg := &config.EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     465,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		FromName: "Test Service",
	}

	// 使用新的API签名：newConnectionPool(cfg, logger)
	pool := newConnectionPool(cfg, &defaultLogger{})
	if pool == nil {
		t.Fatal("newConnectionPool returned nil")
	}

	// 验证连接池创建成功
	if pool.cfg != cfg {
		t.Error("pool.cfg should match input config")
	}

	// 清理
	pool.ClosePool()
}

func TestServiceClose(t *testing.T) {
	cfg := &config.EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     465,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		FromName: "Test Service",
	}

	service := NewService(cfg)
	err := service.Close()
	if err != nil {
		t.Errorf("Close should not return error: %v", err)
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
