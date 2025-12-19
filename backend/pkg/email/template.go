// Package email 邮件发送服务
package email

import (
	"bytes"
	"html/template"
	"sync"
)

// TemplateType 邮件模板类型
type TemplateType string

const (
	TemplateTest              TemplateType = "test"
	TemplateRegisterCode      TemplateType = "register_code"
	TemplateResetPasswordCode TemplateType = "reset_password_code"
	TemplateWelcome           TemplateType = "welcome"
	TemplatePasswordChanged   TemplateType = "password_changed"
	TemplateMembershipExpire  TemplateType = "membership_expire"
	TemplateLoginAlert        TemplateType = "login_alert"
)

// TemplateData 模板数据接口
type TemplateData interface {
	GetSubject() string
	GetData() map[string]interface{}
}

// TemplateManager 邮件模板管理器
type TemplateManager struct {
	templates map[TemplateType]*template.Template
	mu        sync.RWMutex
}

// NewTemplateManager 创建模板管理器
func NewTemplateManager() *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[TemplateType]*template.Template),
	}
	tm.loadDefaultTemplates()
	return tm
}

// loadDefaultTemplates 加载默认模板
func (tm *TemplateManager) loadDefaultTemplates() {
	// 基础布局模板
	baseLayout := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Microsoft YaHei', 'PingFang SC', Arial, sans-serif; background-color: #f5f5f5; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.1); }
        .header { padding: 30px; text-align: center; background: {{.HeaderGradient}}; }
        .header h1 { color: #fff; margin: 0; font-size: 24px; font-weight: 500; }
        .content { padding: 40px 30px; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
        .footer p { margin: 5px 0; }
        @media screen and (max-width: 600px) {
            body { padding: 10px; }
            .content { padding: 20px 15px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.HeaderIcon}} {{.Title}}</h1>
        </div>
        <div class="content">
            {{.Content}}
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复</p>
            <p>© 2024 {{.FromName}} · 保留所有权利</p>
        </div>
    </div>
</body>
</html>`

	// 测试邮件模板
	testTemplate := baseLayout + `
{{define "content"}}
    <div style="text-align: center;">
        <div style="background: #f6ffed; border: 2px solid #b7eb8f; border-radius: 8px; padding: 20px; margin: 20px 0;">
            <p style="color: #52c41a; font-size: 18px; font-weight: bold;">✅ 邮件服务测试成功</p>
        </div>
        <p style="color: #666; margin-top: 20px;">如果您收到这封邮件，说明邮件服务配置正确。</p>
    </div>
{{end}}`

	// 注册验证码模板
	registerCodeTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Microsoft YaHei', 'PingFang SC', Arial, sans-serif; background-color: #f5f5f5; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%); padding: 30px; text-align: center; }
        .header h1 { color: #fff; margin: 0; font-size: 24px; }
        .content { padding: 40px 30px; }
        .code-box { background: linear-gradient(135deg, #f6ffed 0%, #d9f7be 100%); border-radius: 8px; padding: 25px; text-align: center; margin: 20px 0; }
        .code { font-size: 36px; font-weight: bold; color: #52c41a; letter-spacing: 8px; font-family: 'Courier New', monospace; }
        .tip { color: #666; font-size: 14px; margin-top: 20px; line-height: 1.6; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
        @media screen and (max-width: 600px) {
            body { padding: 10px; }
            .content { padding: 20px 15px; }
            .code { font-size: 28px; letter-spacing: 4px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.FromName}}</h1>
        </div>
        <div class="content">
            <p style="color: #333; margin-bottom: 15px;">您好！</p>
            <p style="color: #333; margin-bottom: 20px;">欢迎注册，请使用以下验证码完成注册：</p>
            <div class="code-box">
                <div class="code">{{.Code}}</div>
            </div>
            <p class="tip">验证码有效期为 <strong style="color: #52c41a;">10分钟</strong>，请尽快完成注册。</p>
            <p class="tip" style="color: #999; margin-top: 15px;">如果这不是您本人的操作，请忽略此邮件。</p>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复</p>
            <p>© 2024 {{.FromName}} · 保留所有权利</p>
        </div>
    </div>
</body>
</html>`

	// 密码重置验证码模板
	resetPasswordCodeTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Microsoft YaHei', 'PingFang SC', Arial, sans-serif; background-color: #f5f5f5; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 30px; text-align: center; }
        .header h1 { color: #fff; margin: 0; font-size: 24px; }
        .content { padding: 40px 30px; }
        .code-box { background: linear-gradient(135deg, #f5f7fa 0%, #e4e9f2 100%); border-radius: 8px; padding: 25px; text-align: center; margin: 20px 0; }
        .code { font-size: 36px; font-weight: bold; color: #667eea; letter-spacing: 8px; font-family: 'Courier New', monospace; }
        .tip { color: #666; font-size: 14px; margin-top: 20px; line-height: 1.6; }
        .warning { background: #fff2f0; border: 1px solid #ffccc7; border-radius: 8px; padding: 15px; color: #ff4d4f; font-size: 14px; margin-top: 20px; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
        @media screen and (max-width: 600px) {
            body { padding: 10px; }
            .content { padding: 20px 15px; }
            .code { font-size: 28px; letter-spacing: 4px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.FromName}}</h1>
        </div>
        <div class="content">
            <p style="color: #333; margin-bottom: 15px;">您好！</p>
            <p style="color: #333; margin-bottom: 20px;">您正在进行密码重置操作，请使用以下验证码完成验证：</p>
            <div class="code-box">
                <div class="code">{{.Code}}</div>
            </div>
            <p class="tip">验证码有效期为 <strong style="color: #667eea;">10分钟</strong>，请尽快完成验证。</p>
            <div class="warning">
                <p style="margin: 0;">⚠️ <strong>安全提醒：</strong></p>
                <p style="margin: 10px 0 0 0;">如非本人操作，请忽略此邮件，您的账户安全不会受到影响。</p>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复</p>
            <p>© 2024 {{.FromName}} · 保留所有权利</p>
        </div>
    </div>
</body>
</html>`

	// 欢迎邮件模板
	welcomeTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Microsoft YaHei', 'PingFang SC', Arial, sans-serif; background-color: #f5f5f5; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 40px; text-align: center; }
        .header h1 { color: #fff; margin: 0; font-size: 28px; }
        .header p { color: rgba(255,255,255,0.9); margin-top: 10px; font-size: 16px; }
        .content { padding: 40px 30px; }
        .welcome-box { background: #f8f9fa; border-radius: 8px; padding: 20px; margin: 20px 0; }
        .welcome-box strong { color: #667eea; }
        .welcome-box ul { margin: 15px 0; padding-left: 20px; }
        .welcome-box li { margin: 10px 0; color: #666; line-height: 1.6; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
        @media screen and (max-width: 600px) {
            body { padding: 10px; }
            .content { padding: 20px 15px; }
            .header { padding: 30px 15px; }
            .header h1 { font-size: 24px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 欢迎加入 {{.FromName}}</h1>
            <p>您的账号已创建成功</p>
        </div>
        <div class="content">
            <p style="color: #333; margin-bottom: 15px;">亲爱的 <strong style="color: #667eea;">{{.Username}}</strong>，您好！</p>
            <p style="color: #333; margin-bottom: 20px;">感谢您注册成为我们的用户，您的账号已经创建成功。</p>
            <div class="welcome-box">
                <p><strong>🔐 账户安全提示：</strong></p>
                <ul>
                    <li>请妥善保管您的账号密码</li>
                    <li>建议定期修改密码</li>
                    <li>不要在公共设备上保存登录状态</li>
                    <li>发现异常登录请及时联系我们</li>
                </ul>
            </div>
            <p style="color: #666; margin-top: 20px;">如有任何问题，请随时联系我们。祝您使用愉快！</p>
        </div>
        <div class="footer">
            <p>© 2024 {{.FromName}} · 保留所有权利</p>
        </div>
    </div>
</body>
</html>`

	// 密码修改成功模板
	passwordChangedTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Microsoft YaHei', 'PingFang SC', Arial, sans-serif; background-color: #f5f5f5; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%); padding: 30px; text-align: center; }
        .header h1 { color: #fff; margin: 0; font-size: 24px; }
        .content { padding: 40px 30px; }
        .alert-box { background: #fff7e6; border: 1px solid #ffd591; border-radius: 8px; padding: 15px; margin: 20px 0; }
        .alert-box strong { color: #fa8c16; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
        @media screen and (max-width: 600px) {
            body { padding: 10px; }
            .content { padding: 20px 15px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ 密码修改成功</h1>
        </div>
        <div class="content">
            <p style="color: #333; margin-bottom: 15px;">亲爱的 <strong style="color: #52c41a;">{{.Username}}</strong>，您好！</p>
            <p style="color: #333; margin-bottom: 20px;">您的账户密码已于刚才成功修改。</p>
            <div class="alert-box">
                <p style="margin-bottom: 10px;"><strong>⚠️ 安全提醒：</strong></p>
                <p style="color: #666; margin: 0; line-height: 1.6;">如果这不是您本人的操作，请立即联系客服处理，并检查您的账户安全。</p>
            </div>
            <p style="color: #666; margin-top: 20px;">如有任何问题，请随时联系我们。</p>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复</p>
            <p>© 2024 {{.FromName}} · 保留所有权利</p>
        </div>
    </div>
</body>
</html>`

	// 会员即将到期模板
	membershipExpireTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Microsoft YaHei', 'PingFang SC', Arial, sans-serif; background-color: #f5f5f5; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #faad14 0%, #ffc53d 100%); padding: 30px; text-align: center; }
        .header h1 { color: #fff; margin: 0; font-size: 24px; }
        .content { padding: 40px 30px; text-align: center; }
        .days-box { background: linear-gradient(135deg, #fff7e6 0%, #ffe7ba 100%); border-radius: 12px; padding: 30px; margin: 20px 0; }
        .days { font-size: 48px; font-weight: bold; color: #faad14; font-family: 'Arial Black', sans-serif; }
        .days-text { color: #666; margin-top: 10px; font-size: 16px; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
        @media screen and (max-width: 600px) {
            body { padding: 10px; }
            .content { padding: 20px 15px; }
            .days { font-size: 36px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⏰ 会员即将到期</h1>
        </div>
        <div class="content">
            <p style="color: #333; margin-bottom: 20px;">亲爱的 <strong style="color: #faad14;">{{.Username}}</strong>，您好！</p>
            <div class="days-box">
                <div class="days">{{.DaysLeft}}</div>
                <div class="days-text">天后您的会员将到期</div>
            </div>
            <p style="color: #666; margin-top: 20px;">为了不影响您的使用体验，请及时续费哦~</p>
        </div>
        <div class="footer">
            <p>© 2024 {{.FromName}} · 保留所有权利</p>
        </div>
    </div>
</body>
</html>`

	// 登录提醒模板
	loginAlertTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Microsoft YaHei', 'PingFang SC', Arial, sans-serif; background-color: #f5f5f5; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #1890ff 0%, #40a9ff 100%); padding: 30px; text-align: center; }
        .header h1 { color: #fff; margin: 0; font-size: 24px; }
        .content { padding: 40px 30px; }
        .info-box { background: #f6f8fa; border-radius: 8px; padding: 20px; margin: 20px 0; }
        .info-item { display: flex; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid #eee; }
        .info-item:last-child { border-bottom: none; }
        .label { color: #666; }
        .value { color: #333; font-weight: bold; }
        .alert-box { background: #fff2f0; border: 1px solid #ffccc7; border-radius: 8px; padding: 15px; margin: 20px 0; }
        .alert-box strong { color: #ff4d4f; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #999; font-size: 12px; }
        @media screen and (max-width: 600px) {
            body { padding: 10px; }
            .content { padding: 20px 15px; }
            .info-item { flex-direction: column; gap: 5px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔔 账户登录提醒</h1>
        </div>
        <div class="content">
            <p style="color: #333; margin-bottom: 15px;">亲爱的 <strong style="color: #1890ff;">{{.Username}}</strong>，您好！</p>
            <p style="color: #333; margin-bottom: 20px;">您的账户刚刚在新设备上登录，登录信息如下：</p>
            <div class="info-box">
                <div class="info-item">
                    <span class="label">登录IP</span>
                    <span class="value">{{.IP}}</span>
                </div>
                <div class="info-item">
                    <span class="label">登录地点</span>
                    <span class="value">{{.Location}}</span>
                </div>
                <div class="info-item">
                    <span class="label">登录设备</span>
                    <span class="value">{{.Device}}</span>
                </div>
            </div>
            <div class="alert-box">
                <p style="margin: 0; line-height: 1.6;"><strong>⚠️ 安全提醒：</strong> 如果这不是您本人的操作，请立即修改密码并检查账户安全。</p>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复</p>
            <p>© 2024 {{.FromName}} · 保留所有权利</p>
        </div>
    </div>
</body>
</html>`

	// 注册模板
	tm.RegisterTemplate(TemplateTest, testTemplate)
	tm.RegisterTemplate(TemplateRegisterCode, registerCodeTemplate)
	tm.RegisterTemplate(TemplateResetPasswordCode, resetPasswordCodeTemplate)
	tm.RegisterTemplate(TemplateWelcome, welcomeTemplate)
	tm.RegisterTemplate(TemplatePasswordChanged, passwordChangedTemplate)
	tm.RegisterTemplate(TemplateMembershipExpire, membershipExpireTemplate)
	tm.RegisterTemplate(TemplateLoginAlert, loginAlertTemplate)
}

// RegisterTemplate 注册模板
func (tm *TemplateManager) RegisterTemplate(name TemplateType, content string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tmpl, err := template.New(string(name)).Parse(content)
	if err != nil {
		return err
	}

	tm.templates[name] = tmpl
	return nil
}

// Render 渲染模板
func (tm *TemplateManager) Render(name TemplateType, data interface{}) (string, error) {
	tm.mu.RLock()
	tmpl, exists := tm.templates[name]
	tm.mu.RUnlock()

	if !exists {
		return "", ErrTemplateNotFound
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// HasTemplate 检查模板是否存在
func (tm *TemplateManager) HasTemplate(name TemplateType) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	_, exists := tm.templates[name]
	return exists
}
