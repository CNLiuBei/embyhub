package email

import "errors"

var (
	// ErrTemplateNotFound 模板未找到
	ErrTemplateNotFound = errors.New("email template not found")

	// ErrEmailNotEnabled 邮件服务未启用
	ErrEmailNotEnabled = errors.New("email service is not enabled")

	// ErrInvalidRecipient 无效的收件人
	ErrInvalidRecipient = errors.New("invalid email recipient")

	// ErrInvalidConfig 无效的配置
	ErrInvalidConfig = errors.New("invalid email configuration")

	// ErrSendFailed 发送失败
	ErrSendFailed = errors.New("failed to send email")

	// ErrConnectionFailed 连接失败
	ErrConnectionFailed = errors.New("failed to connect to email server")

	// ErrAuthFailed 认证失败
	ErrAuthFailed = errors.New("email authentication failed")
)
