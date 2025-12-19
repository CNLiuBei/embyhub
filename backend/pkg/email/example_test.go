package email_test

import (
	"fmt"
	"log"
	"time"

	"feiniu-user-system/internal/config"
	"feiniu-user-system/pkg/email"
)

// Example_basicUsage 基本使用示例
func Example_basicUsage() {
	cfg := &config.EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     465,
		Username: "sender@example.com",
		Password: "password",
		From:     "sender@example.com",
		FromName: "My Service",
	}

	service := email.NewService(cfg)
	defer service.Close()

	// 发送测试邮件
	err := service.SendTestEmail("recipient@example.com")
	if err != nil {
		log.Printf("Failed to send test email: %v", err)
	}
}

// Example_registerCode 发送注册验证码
func Example_registerCode() {
	cfg := &config.EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     465,
		Username: "sender@example.com",
		Password: "password",
		From:     "sender@example.com",
		FromName: "My Service",
	}

	service := email.NewService(cfg)
	defer service.Close()

	// 发送注册验证码
	code := "123456"
	err := service.SendRegisterCode("user@example.com", code)
	if err != nil {
		log.Printf("Failed to send register code: %v", err)
	}
}

// Example_withQueue 使用异步队列
func Example_withQueue() {
	cfg := &config.EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     465,
		Username: "sender@example.com",
		Password: "password",
		From:     "sender@example.com",
		FromName: "My Service",
	}

	service := email.NewService(cfg)
	defer service.Close()

	// 创建队列：2个工作协程，队列大小100
	queue := email.NewQueue(service, 2, 100)
	defer queue.Stop(10 * time.Second)

	// 异步发送邮件
	for i := 0; i < 10; i++ {
		recipient := fmt.Sprintf("user%d@example.com", i)
		err := queue.EnqueueHTML(
			recipient,
			"Welcome",
			"<h1>Welcome to our service!</h1>",
		)
		if err != nil {
			log.Printf("Failed to enqueue email: %v", err)
		}
	}

	// 等待邮件发送
	time.Sleep(2 * time.Second)

	// 查看指标
	metrics := queue.GetMetrics()
	fmt.Printf("Sent: %d, Failed: %d\n", metrics.TotalSent, metrics.TotalFailed)
}

// Example_customTemplate 使用自定义模板
func Example_customTemplate() {
	cfg := &config.EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     465,
		Username: "sender@example.com",
		Password: "password",
		From:     "sender@example.com",
		FromName: "My Service",
	}

	service := email.NewService(cfg)
	defer service.Close()

	// 使用欢迎邮件模板
	data := map[string]interface{}{
		"FromName": "My Service",
		"Username": "John Doe",
	}

	err := service.SendTemplate(
		"user@example.com",
		"Welcome!",
		email.TemplateWelcome,
		data,
	)
	if err != nil {
		log.Printf("Failed to send welcome email: %v", err)
	}
}

// Example_loginAlert 发送登录提醒
func Example_loginAlert() {
	cfg := &config.EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     465,
		Username: "sender@example.com",
		Password: "password",
		From:     "sender@example.com",
		FromName: "My Service",
	}

	service := email.NewService(cfg)
	defer service.Close()

	err := service.SendLoginAlert(
		"user@example.com",
		"John Doe",
		"192.168.1.100",
		"北京市",
		"Chrome on Windows",
	)
	if err != nil {
		log.Printf("Failed to send login alert: %v", err)
	}
}

// Example_monitoring 监控队列
func Example_monitoring() {
	cfg := &config.EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     465,
		Username: "sender@example.com",
		Password: "password",
		From:     "sender@example.com",
		FromName: "My Service",
	}

	service := email.NewService(cfg)
	defer service.Close()

	queue := email.NewQueue(service, 2, 100)
	defer queue.Stop(10 * time.Second)

	// 启动监控协程
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			metrics := queue.GetMetrics()
			log.Printf("Queue metrics:")
			log.Printf("  Total enqueued: %d", metrics.TotalEnqueued)
			log.Printf("  Total sent: %d", metrics.TotalSent)
			log.Printf("  Total failed: %d", metrics.TotalFailed)
			log.Printf("  Current queue: %d", metrics.CurrentQueue)
			log.Printf("  Healthy: %v", queue.IsHealthy())
		}
	}()

	// 发送一些邮件
	for i := 0; i < 5; i++ {
		queue.EnqueueHTML(
			fmt.Sprintf("user%d@example.com", i),
			"Test",
			"<p>Test email</p>",
		)
	}

	time.Sleep(1 * time.Minute)
}

// Example_batchSend 批量发送
func Example_batchSend() {
	cfg := &config.EmailConfig{
		Enabled:  true,
		Host:     "smtp.example.com",
		Port:     465,
		Username: "sender@example.com",
		Password: "password",
		From:     "sender@example.com",
		FromName: "My Service",
	}

	service := email.NewService(cfg)
	defer service.Close()

	// 创建较大的队列用于批量发送
	queue := email.NewQueue(service, 4, 500)
	defer queue.Stop(30 * time.Second)

	// 批量发送营销邮件
	recipients := []string{
		"user1@example.com",
		"user2@example.com",
		"user3@example.com",
		// ... 更多收件人
	}

	subject := "Monthly Newsletter"
	body := `
<!DOCTYPE html>
<html>
<body>
    <h1>Monthly Newsletter</h1>
    <p>Check out our latest updates...</p>
</body>
</html>`

	successCount := 0
	for _, recipient := range recipients {
		err := queue.EnqueueHTML(recipient, subject, body)
		if err != nil {
			log.Printf("Failed to enqueue for %s: %v", recipient, err)
		} else {
			successCount++
		}
	}

	log.Printf("Enqueued %d emails", successCount)

	// 等待所有邮件发送完成
	time.Sleep(10 * time.Second)

	metrics := queue.GetMetrics()
	log.Printf("Final metrics: %d sent, %d failed",
		metrics.TotalSent, metrics.TotalFailed)
}
