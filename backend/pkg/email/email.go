package email

import (
	"fmt"
	"strconv"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dm "github.com/alibabacloud-go/dm-20151123/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/resend/resend-go/v2"
	"gopkg.in/gomail.v2"
)

// Config 邮件配置
type Config struct {
	// 通用配置
	Provider string // smtp, resend, aliyun
	From     string
	FromName string

	// SMTP配置
	Host     string
	Port     int
	User     string
	Password string
	UseSSL   bool

	// Resend配置
	ResendAPIKey string

	// 阿里云配置
	AliyunAccessKeyID     string
	AliyunAccessKeySecret string
	AliyunRegion          string // 如 cn-hangzhou
}

// Client 邮件客户端
type Client struct {
	config *Config
}

// NewClient 创建邮件客户端
func NewClient(config *Config) *Client {
	return &Client{config: config}
}

// Send 发送邮件（根据Provider配置选择发送方式）
func (c *Client) Send(to, subject, body string) error {
	switch c.config.Provider {
	case "aliyun":
		return c.sendWithAliyun(to, subject, body)
	case "resend":
		return c.sendWithResend(to, subject, body)
	case "aliyun_smtp", "smtp":
		return c.sendWithSMTP(to, subject, body)
	default:
		return c.sendWithSMTP(to, subject, body)
	}
}

// sendWithResend 使用Resend发送邮件
func (c *Client) sendWithResend(to, subject, body string) error {
	client := resend.NewClient(c.config.ResendAPIKey)

	// 构建发件人
	from := c.config.From
	if c.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", c.config.FromName, c.config.From)
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Html:    body,
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("Resend发送失败: %w", err)
	}
	return nil
}

// sendWithSMTP 使用SMTP发送邮件
func (c *Client) sendWithSMTP(to, subject, body string) error {
	if c.config.Host == "" || c.config.User == "" {
		return fmt.Errorf("SMTP未配置")
	}

	m := gomail.NewMessage(gomail.SetEncoding(gomail.Base64))

	// 设置发件人
	if c.config.FromName != "" {
		m.SetAddressHeader("From", c.config.From, c.config.FromName)
	} else {
		m.SetHeader("From", c.config.From)
	}

	m.SetAddressHeader("To", to, "")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(c.config.Host, c.config.Port, c.config.User, c.config.Password)
	d.SSL = c.config.UseSSL

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("SMTP发送失败: %w", err)
	}
	return nil
}

// sendWithAliyun 使用阿里云邮件推送
func (c *Client) sendWithAliyun(to, subject, body string) error {
	config := &openapi.Config{
		AccessKeyId:     tea.String(c.config.AliyunAccessKeyID),
		AccessKeySecret: tea.String(c.config.AliyunAccessKeySecret),
	}

	region := c.config.AliyunRegion
	if region == "" {
		region = "cn-hangzhou"
	}
	config.Endpoint = tea.String(fmt.Sprintf("dm.%s.aliyuncs.com", region))

	client, err := dm.NewClient(config)
	if err != nil {
		return fmt.Errorf("创建阿里云客户端失败: %w", err)
	}

	request := &dm.SingleSendMailRequest{
		AccountName:    tea.String(c.config.From),
		AddressType:    tea.Int32(1),
		ReplyToAddress: tea.Bool(false),
		ToAddress:      tea.String(to),
		Subject:        tea.String(subject),
		HtmlBody:       tea.String(body),
	}

	if c.config.FromName != "" {
		request.FromAlias = tea.String(c.config.FromName)
	}

	_, err = client.SingleSendMail(request)
	if err != nil {
		return fmt.Errorf("阿里云发送失败: %w", err)
	}
	return nil
}

// SendPasswordResetCode 发送密码重置验证码
func (c *Client) SendPasswordResetCode(to, code string) error {
	subject, body := PasswordResetEmail(code)
	return c.Send(to, subject, body)
}

// SendVerificationCode 发送验证码邮件
func (c *Client) SendVerificationCode(to, code string) error {
	subject, body := VerificationCodeEmail(code, "账号注册")
	return c.Send(to, subject, body)
}

// SendWelcomeEmail 发送欢迎邮件
func (c *Client) SendWelcomeEmail(to, username string) error {
	subject, body := WelcomeEmail(username)
	return c.Send(to, subject, body)
}

// SendVipExpiringEmail 发送VIP到期提醒
func (c *Client) SendVipExpiringEmail(to, username, expireDate string, daysLeft int) error {
	subject, body := VipExpiringEmail(username, expireDate, daysLeft)
	return c.Send(to, subject, body)
}

// SendLoginAlertEmail 发送登录提醒
func (c *Client) SendLoginAlertEmail(to, username, ip, device, loginTime string) error {
	subject, body := LoginAlertEmail(username, ip, device, loginTime)
	return c.Send(to, subject, body)
}

// SendPasswordChangedEmail 发送密码修改通知
func (c *Client) SendPasswordChangedEmail(to, username, changeTime string) error {
	subject, body := PasswordChangedEmail(username, changeTime)
	return c.Send(to, subject, body)
}

// SendTestEmail 发送测试邮件
func (c *Client) SendTestEmail(to string) error {
	subject, body := TestEmail()
	return c.Send(to, subject, body)
}

// base64Encode Base64编码
func base64Encode(s string) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	data := []byte(s)
	result := make([]byte, 0, (len(data)+2)/3*4)

	for i := 0; i < len(data); i += 3 {
		var n uint32
		remaining := len(data) - i

		n = uint32(data[i]) << 16
		if remaining > 1 {
			n |= uint32(data[i+1]) << 8
		}
		if remaining > 2 {
			n |= uint32(data[i+2])
		}

		result = append(result, base64Chars[(n>>18)&0x3F])
		result = append(result, base64Chars[(n>>12)&0x3F])
		if remaining > 1 {
			result = append(result, base64Chars[(n>>6)&0x3F])
		} else {
			result = append(result, '=')
		}
		if remaining > 2 {
			result = append(result, base64Chars[n&0x3F])
		} else {
			result = append(result, '=')
		}
	}

	return string(result)
}

// GetConfigFromDB 从数据库获取邮件配置
func GetConfigFromDB(configs map[string]string) *Config {
	port, _ := strconv.Atoi(configs["smtp_port"])
	if port == 0 {
		port = 465
	}

	useSSL := strings.ToLower(configs["smtp_ssl"]) == "true"

	// 兼容旧配置：优先使用新的email_from，否则使用smtp_from
	from := configs["email_from"]
	if from == "" {
		from = configs["smtp_from"]
	}
	fromName := configs["email_from_name"]
	if fromName == "" {
		fromName = configs["smtp_from_name"]
	}

	return &Config{
		Provider:              configs["email_provider"],
		From:                  from,
		FromName:              fromName,
		Host:                  configs["smtp_host"],
		Port:                  port,
		User:                  configs["smtp_user"],
		Password:              configs["smtp_password"],
		UseSSL:                useSSL,
		ResendAPIKey:          configs["resend_api_key"],
		AliyunAccessKeyID:     configs["aliyun_access_key_id"],
		AliyunAccessKeySecret: configs["aliyun_access_key_secret"],
		AliyunRegion:          configs["aliyun_region"],
	}
}

// GetConfigKeys 获取所有邮件配置键名
func GetConfigKeys() []string {
	return []string{
		"email_provider", "email_from", "email_from_name",
		"smtp_host", "smtp_port", "smtp_user", "smtp_password", "smtp_ssl",
		"smtp_from", "smtp_from_name", // 兼容旧配置
		"resend_api_key",
		"aliyun_access_key_id", "aliyun_access_key_secret", "aliyun_region",
	}
}
