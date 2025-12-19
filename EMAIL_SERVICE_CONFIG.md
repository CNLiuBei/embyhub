# 邮件服务配置说明

## 📧 功能说明

邮件服务用于发送：
- ✉️ 注册验证码
- 🔑 密码重置验证码  
- 👋 注册成功欢迎邮件
- 🔒 密码修改通知
- ⏰ 会员到期提醒
- 🌍 异地登录提醒
- 🧪 测试邮件

## 🔧 配置结构

### 后端配置（数据库存储）

```json
{
  "enabled": true,                    // 是否启用邮件服务
  "host": "smtp.example.com",         // SMTP服务器地址
  "port": 587,                        // SMTP端口
  "username": "your_email@domain.com", // SMTP账号
  "password": "授权码",                // SMTP授权码（不是邮箱密码）
  "from": "your_email@domain.com",     // 发件邮箱
  "from_name": "飞牛影视"              // 发件人名称
}
```

### 配置文件（默认配置）

`/vol1/1000/FnHub/feiniu-user-system/backend/config/config.yaml`:

```yaml
email:
  enabled: true
  host: "smtp.qq.com"
  port: 587
  username: "your_email@qq.com"
  password: "your_smtp_password"
  from: "your_email@qq.com"
  from_name: "飞牛影视"
```

## 📮 常用邮箱配置参数

### QQ邮箱

```yaml
host: "smtp.qq.com"
port: 587          # 或 465 (SSL)
username: "你的QQ邮箱@qq.com"
password: "授权码"  # 不是QQ密码！
from: "你的QQ邮箱@qq.com"
from_name: "飞牛影视"
```

**获取QQ邮箱授权码**：
1. 登录QQ邮箱网页版
2. 设置 → 账户 → POP3/IMAP/SMTP/Exchange/CardDAV/CalDAV服务
3. 开启"POP3/SMTP服务"
4. 生成授权码（不是你的QQ密码）
5. 保存授权码用于配置

### 163邮箱

```yaml
host: "smtp.163.com"
port: 587          # 或 465 (SSL)
username: "你的163邮箱@163.com"
password: "授权码"  # 不是登录密码
from: "你的163邮箱@163.com"
from_name: "飞牛影视"
```

**获取163邮箱授权码**：
1. 登录163邮箱
2. 设置 → POP3/SMTP/IMAP
3. 开启"SMTP服务"
4. 新增授权码
5. 使用生成的授权码

### 126邮箱

```yaml
host: "smtp.126.com"
port: 587
username: "你的126邮箱@126.com"
password: "授权码"
from: "你的126邮箱@126.com"
from_name: "飞牛影视"
```

### Gmail

```yaml
host: "smtp.gmail.com"
port: 587          # 或 465 (SSL)
username: "your_email@gmail.com"
password: "应用专用密码"  # 不是Gmail登录密码
from: "your_email@gmail.com"
from_name: "飞牛影视"
```

**获取Gmail应用专用密码**：
1. 启用两步验证
2. Google账户 → 安全性 → 两步验证 → 应用专用密码
3. 生成应用专用密码
4. 使用生成的16位密码

### 阿里云企业邮箱

```yaml
host: "smtp.mxhichina.com"
port: 465          # SSL端口
username: "user@your-domain.com"
password: "邮箱密码"
from: "user@your-domain.com"
from_name: "飞牛影视"
```

### 腾讯企业邮箱

```yaml
host: "smtp.exmail.qq.com"
port: 465          # 或 587
username: "user@your-domain.com"
password: "邮箱密码或授权码"
from: "user@your-domain.com"
from_name: "飞牛影视"
```

### Office 365 / Outlook

```yaml
host: "smtp.office365.com"
port: 587
username: "your_email@outlook.com"
password: "邮箱密码"
from: "your_email@outlook.com"
from_name: "飞牛影视"
```

## 🔌 端口说明

| 端口 | 协议 | 说明 |
|------|------|------|
| 25 | SMTP | 标准SMTP端口（可能被运营商封禁） |
| 587 | SMTP+STARTTLS | 推荐使用（加密传输） |
| 465 | SMTP+SSL | SSL加密端口 |

**推荐使用**：
- ✅ **587端口** - 支持STARTTLS加密，兼容性好
- ✅ **465端口** - SSL加密，安全性高

## ⚙️ 配置步骤

### 步骤1：选择邮箱服务商

推荐优先级：
1. ⭐⭐⭐ **QQ邮箱** - 配置简单，免费，稳定
2. ⭐⭐⭐ **163邮箱** - 免费，稳定
3. ⭐⭐ **Gmail** - 需要国际网络
4. ⭐⭐⭐⭐ **企业邮箱** - 专业，但需要付费

### 步骤2：获取SMTP授权码

⚠️ **重要**：不是邮箱登录密码！

各邮箱服务商都会提供"授权码"或"应用专用密码"，用于第三方应用登录。

### 步骤3：在系统中配置

1. 访问：`http://localhost:3000/admin/settings`
2. 进入"邮件服务"标签
3. 填写配置信息：
   - SMTP服务器地址：如 `smtp.qq.com`
   - 端口：`587` 或 `465`
   - 用户名邮箱：你的邮箱地址
   - 邮箱授权码：获取的授权码（不是密码）
   - 发件人邮箱：同用户名邮箱
   - 发件人名称：如"飞牛影视"
4. 打开"启用邮件评估"开关
5. 点击"保存设置"

### 步骤4：发送测试邮件

1. 在"测试邮件"部分
2. 输入收件人邮箱
3. 点击"发送测试"
4. 检查邮箱是否收到测试邮件

## 🐛 常见问题

### 问题1：535 Authentication Failed

**原因**：
- 使用了邮箱登录密码而不是授权码
- 授权码错误
- 未开启SMTP服务

**解决**：
1. 确认使用的是授权码，不是登录密码
2. 重新生成授权码
3. 确认已开启SMTP服务

### 问题2：connect: connection refused

**原因**：
- SMTP服务器地址错误
- 端口被防火墙封锁
- 网络连接问题

**解决**：
1. 检查SMTP服务器地址是否正确
2. 尝试其他端口（587/465）
3. 检查服务器防火墙设置

### 问题3：554 DT:SPM

**原因**：
- 邮件被识别为垃圾邮件
- 发送频率过高

**解决**：
1. 等待一段时间后重试
2. 检查邮件内容，避免敏感词
3. 联系邮件服务商解除限制

### 问题4：邮件进入垃圾箱

**原因**：
- SPF、DKIM配置缺失
- 发件人信任度低

**解决**：
1. 使用企业邮箱
2. 配置SPF记录（需域名管理权限）
3. 避免发送过多营销类内容

## 🔒 安全建议

### 1. 使用授权码，不用密码

✅ **正确**：使用SMTP授权码  
❌ **错误**：使用邮箱登录密码

### 2. 配置存储安全

- 授权码加密存储在数据库
- 不在日志中打印授权码
- 定期更换授权码

### 3. 发送频率控制

- 限制每分钟发送数量
- 避免短时间大量发送
- 防止被邮件服务商封禁

## 📊 监控和日志

### 发送成功日志

```
[INFO] Sending register code to=user@example.com code=123456
[INFO] Register code sent successfully to=user@example.com
```

### 发送失败日志

```
[ERROR] Failed to send email attempt=1 error=535 Authentication Failed
[INFO] Retrying email send attempt=2 wait=1s
```

## 🧪 测试配置

### 测试邮件内容

```html
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px;">
    <div style="max-width: 500px; margin: 0 auto; background: #f5f5f5; 
         border-radius: 10px; padding: 30px;">
        <h2 style="color: #333; text-align: center;">
            ✅ 邮件服务测试成功
        </h2>
        <p style="color: #666; text-align: center;">
            如果您收到这封邮件，说明邮件服务配置正确。
        </p>
    </div>
</body>
</html>
```

### 测试API

```bash
# 发送测试邮件
curl -X POST http://localhost:8080/api/v1/admin/settings/email/test \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"to": "test@example.com"}'
```

## 🎯 快速配置示例

### QQ邮箱（推荐新手）

```json
{
  "enabled": true,
  "host": "smtp.qq.com",
  "port": 587,
  "username": "123456789@qq.com",
  "password": "abcdefghijklmnop",  // 授权码，16位
  "from": "123456789@qq.com",
  "from_name": "飞牛影视"
}
```

### 163邮箱

```json
{
  "enabled": true,
  "host": "smtp.163.com",
  "port": 587,
  "username": "your_email@163.com",
  "password": "XXXXXXXXXXXXXXXX",  // 授权码
  "from": "your_email@163.com",
  "from_name": "飞牛影视"
}
```

## 📝 配置检查清单

- [ ] 选择邮箱服务商
- [ ] 开启SMTP服务
- [ ] 获取授权码（不是密码）
- [ ] 填写正确的SMTP服务器地址
- [ ] 选择正确的端口（587或465）
- [ ] 填写用户名（完整邮箱地址）
- [ ] 填写授权码
- [ ] 填写发件人邮箱
- [ ] 填写发件人名称
- [ ] 启用邮件服务开关
- [ ] 保存配置
- [ ] 发送测试邮件验证

## 🎉 总结

邮件服务配置的关键：

1. **获取授权码**（不是密码）
2. **正确的SMTP服务器地址和端口**
3. **完整的邮箱地址作为用户名**
4. **测试验证配置是否正确**

推荐使用QQ邮箱或163邮箱进行配置，简单易用，稳定可靠！
