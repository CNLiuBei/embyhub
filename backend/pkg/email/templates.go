package email

import "fmt"

// 邮件模板基础样式 - 现代卡片风格
const baseTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 100%%); font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
    <table width="100%%" cellpadding="0" cellspacing="0" style="min-height: 100vh;">
        <tr>
            <td align="center" style="padding: 40px 20px;">
                <table width="100%%" style="max-width: 520px; background: #ffffff; border-radius: 16px; overflow: hidden; box-shadow: 0 20px 60px rgba(0,0,0,0.3);">
                    <!-- Logo区域 -->
                    <tr>
                        <td style="background: linear-gradient(135deg, %s 0%%, %s 100%%); padding: 40px 30px; text-align: center;">
                            <div style="width: 60px; height: 60px; background: rgba(255,255,255,0.2); border-radius: 50%%; margin: 0 auto 15px; display: flex; align-items: center; justify-content: center;">
                                <span style="font-size: 28px;">%s</span>
                            </div>
                            <h1 style="color: #ffffff; margin: 0; font-size: 24px; font-weight: 600; letter-spacing: 1px;">%s</h1>
                        </td>
                    </tr>
                    <!-- 内容区域 -->
                    <tr>
                        <td style="padding: 40px 35px;">
                            %s
                        </td>
                    </tr>
                    <!-- 底部区域 -->
                    <tr>
                        <td style="background: #f8f9fa; padding: 25px 35px; border-top: 1px solid #eee;">
                            <table width="100%%" cellpadding="0" cellspacing="0">
                                <tr>
                                    <td style="text-align: center;">
                                        <p style="margin: 0 0 8px; color: #1890ff; font-weight: 600; font-size: 14px;">Emby 用户管理系统</p>
                                        <p style="margin: 0; color: #999; font-size: 12px;">此邮件由系统自动发送，请勿直接回复</p>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                </table>
                <!-- 版权信息 -->
                <p style="margin-top: 30px; color: rgba(255,255,255,0.5); font-size: 12px;">
                    © 2024 Emby Hub. All rights reserved.
                </p>
            </td>
        </tr>
    </table>
</body>
</html>
`

// WelcomeEmail 欢迎邮件
func WelcomeEmail(username string) (subject, body string) {
	subject = "🎉 欢迎加入 Emby Hub"
	content := fmt.Sprintf(`
        <h2 style="margin: 0 0 20px; color: #1a1a2e; font-size: 22px;">Hi，%s 👋</h2>
        <p style="color: #555; line-height: 1.8; margin: 0 0 25px;">
            欢迎加入 Emby 用户管理系统！您的账号已成功创建，Emby 服务账号也已同步开通。
        </p>
        <div style="background: linear-gradient(135deg, #e8f5e9 0%%, #c8e6c9 100%%); border-radius: 12px; padding: 20px; margin: 25px 0;">
            <p style="margin: 0; color: #2e7d32; font-size: 14px;">
                ✨ <strong>快速开始</strong><br><br>
                • 使用相同账号密码登录 Emby 客户端<br>
                • 建议尽快绑定邮箱以便找回密码
            </p>
        </div>
        <p style="color: #888; font-size: 13px; margin: 0;">祝您观影愉快！🎬</p>
    `, username)
	body = fmt.Sprintf(baseTemplate, "#00c853", "#69f0ae", "🎊", "欢迎加入", content)
	return
}

// VerificationCodeEmail 验证码邮件
func VerificationCodeEmail(code, purpose string) (subject, body string) {
	subject = "📧 您的验证码"
	content := fmt.Sprintf(`
        <p style="color: #555; line-height: 1.6; margin: 0 0 25px;">您正在进行 <strong style="color: #667eea;">%s</strong> 操作，请使用以下验证码：</p>
        <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); border-radius: 12px; padding: 30px; text-align: center; margin: 25px 0;">
            <span style="font-size: 42px; font-weight: bold; color: #fff; letter-spacing: 12px; text-shadow: 0 2px 4px rgba(0,0,0,0.2);">%s</span>
        </div>
        <div style="background: #fff3e0; border-radius: 8px; padding: 15px; margin: 20px 0;">
            <p style="margin: 0; color: #e65100; font-size: 13px;">
                ⏱️ 验证码 <strong>10分钟</strong> 内有效<br>
                🔒 请勿将验证码泄露给任何人
            </p>
        </div>
    `, purpose, code)
	body = fmt.Sprintf(baseTemplate, "#667eea", "#764ba2", "🔐", "验证码", content)
	return
}

// PasswordResetEmail 密码重置邮件
func PasswordResetEmail(code string) (subject, body string) {
	subject = "🔑 密码重置验证码"
	content := fmt.Sprintf(`
        <p style="color: #555; line-height: 1.6; margin: 0 0 25px;">您正在重置账号密码，请使用以下验证码完成操作：</p>
        <div style="background: linear-gradient(135deg, #ff5252 0%%, #ff1744 100%%); border-radius: 12px; padding: 30px; text-align: center; margin: 25px 0;">
            <span style="font-size: 42px; font-weight: bold; color: #fff; letter-spacing: 12px; text-shadow: 0 2px 4px rgba(0,0,0,0.2);">%s</span>
        </div>
        <div style="background: #ffebee; border-radius: 8px; padding: 15px; margin: 20px 0;">
            <p style="margin: 0; color: #c62828; font-size: 13px;">
                ⚠️ 如非本人操作，请忽略此邮件<br>
                🛡️ 请立即检查您的账号安全
            </p>
        </div>
    `, code)
	body = fmt.Sprintf(baseTemplate, "#ff5252", "#ff1744", "🔑", "密码重置", content)
	return
}

// VipExpiringEmail VIP即将到期提醒
func VipExpiringEmail(username string, expireDate string, daysLeft int) (subject, body string) {
	subject = "⏰ VIP会员即将到期"
	content := fmt.Sprintf(`
        <h2 style="margin: 0 0 20px; color: #1a1a2e; font-size: 20px;">亲爱的 %s</h2>
        <p style="color: #555; line-height: 1.6; margin: 0 0 25px;">您的VIP会员即将到期，请注意续费时间：</p>
        <div style="background: linear-gradient(135deg, #fff8e1 0%%, #ffecb3 100%%); border-radius: 12px; padding: 25px; margin: 25px 0; text-align: center;">
            <p style="margin: 0 0 10px; color: #f57c00; font-size: 14px;">到期时间</p>
            <p style="margin: 0 0 15px; color: #e65100; font-size: 24px; font-weight: bold;">%s</p>
            <div style="display: inline-block; background: #ff5722; color: #fff; padding: 8px 20px; border-radius: 20px; font-weight: bold;">
                剩余 %d 天
            </div>
        </div>
        <p style="color: #888; font-size: 13px; margin: 0;">及时续费，畅享无限精彩内容 🎬</p>
    `, username, expireDate, daysLeft)
	body = fmt.Sprintf(baseTemplate, "#ff9800", "#ffc107", "👑", "VIP提醒", content)
	return
}

// LoginAlertEmail 异常登录提醒
func LoginAlertEmail(username, ip, device, loginTime string) (subject, body string) {
	subject = "🚨 账号登录提醒"
	content := fmt.Sprintf(`
        <h2 style="margin: 0 0 20px; color: #1a1a2e; font-size: 20px;">安全提醒</h2>
        <p style="color: #555; line-height: 1.6; margin: 0 0 25px;">您的账号 <strong>%s</strong> 刚刚进行了登录：</p>
        <div style="background: #f5f5f5; border-radius: 12px; padding: 20px; margin: 25px 0;">
            <table style="width: 100%%; border-collapse: collapse;">
                <tr><td style="padding: 8px 0; color: #888; width: 80px;">🕐 时间</td><td style="color: #333;">%s</td></tr>
                <tr><td style="padding: 8px 0; color: #888;">🌐 IP</td><td style="color: #333;">%s</td></tr>
                <tr><td style="padding: 8px 0; color: #888;">💻 设备</td><td style="color: #333; word-break: break-all;">%s</td></tr>
            </table>
        </div>
        <div style="background: #ffebee; border-radius: 8px; padding: 15px; margin: 20px 0;">
            <p style="margin: 0; color: #c62828; font-size: 13px;">
                ⚠️ 如非本人操作，请立即修改密码！
            </p>
        </div>
    `, username, loginTime, ip, device)
	body = fmt.Sprintf(baseTemplate, "#f44336", "#e53935", "🛡️", "安全提醒", content)
	return
}

// PasswordChangedEmail 密码修改通知
func PasswordChangedEmail(username, changeTime string) (subject, body string) {
	subject = "✅ 密码修改成功"
	content := fmt.Sprintf(`
        <div style="text-align: center; margin-bottom: 25px;">
            <div style="width: 70px; height: 70px; background: linear-gradient(135deg, #4caf50 0%%, #8bc34a 100%%); border-radius: 50%%; margin: 0 auto 15px; display: flex; align-items: center; justify-content: center;">
                <span style="font-size: 32px;">✓</span>
            </div>
            <h2 style="margin: 0; color: #1a1a2e; font-size: 20px;">密码已更新</h2>
        </div>
        <p style="color: #555; line-height: 1.6; text-align: center; margin: 0 0 25px;">
            账号 <strong>%s</strong> 的密码已于<br><strong>%s</strong> 成功修改
        </p>
        <div style="background: #fff3e0; border-radius: 8px; padding: 15px; margin: 20px 0;">
            <p style="margin: 0; color: #e65100; font-size: 13px; text-align: center;">
                🔐 如非本人操作，请立即联系管理员
            </p>
        </div>
    `, username, changeTime)
	body = fmt.Sprintf(baseTemplate, "#4caf50", "#8bc34a", "🔒", "密码修改", content)
	return
}

// TestEmail 测试邮件
func TestEmail() (subject, body string) {
	subject = "✅ 邮件服务配置成功"
	content := `
        <div style="text-align: center;">
            <div style="width: 80px; height: 80px; background: linear-gradient(135deg, #4caf50 0%, #8bc34a 100%); border-radius: 50%; margin: 0 auto 20px; display: flex; align-items: center; justify-content: center;">
                <span style="font-size: 40px;">✓</span>
            </div>
            <h2 style="margin: 0 0 15px; color: #1a1a2e; font-size: 22px;">配置成功！</h2>
            <p style="color: #555; line-height: 1.6; margin: 0;">
                您的SMTP邮件服务已正确配置<br>
                系统可以正常发送各类通知邮件
            </p>
        </div>
    `
	body = fmt.Sprintf(baseTemplate, "#4caf50", "#8bc34a", "📧", "测试成功", content)
	return
}
