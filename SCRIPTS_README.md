# 飞牛用户系统 - 启动脚本说明

## 📋 脚本列表

### 1. 一键启动/停止脚本（推荐）

#### `start-all.sh` - 启动所有服务
一键启动前端和后端服务。

```bash
./start-all.sh
```

**功能：**
- ✅ 自动检查并停止旧服务
- ✅ 启动后端服务（端口8080）
- ✅ 启动前端服务（端口3000）
- ✅ 显示服务状态和访问地址

#### `stop-all.sh` - 停止所有服务
停止前端和后端服务。

```bash
./stop-all.sh
```

**功能：**
- ✅ 停止后端服务（端口8080）
- ✅ 停止前端服务（端口3000）
- ✅ 清理相关进程

---

### 2. 单独启动脚本

#### `backend/start.sh` - 后端启动脚本
单独启动后端服务。

```bash
cd backend
./start.sh
```

**功能：**
- ✅ 检查Go环境
- ✅ 自动检测并停止占用端口8080的服务
- ✅ 启动后端服务
- ✅ 记录PID和日志
- ✅ 显示初始化日志（包括超级管理员创建和飞牛同步）

**日志位置：** `backend/logs/backend_YYYYMMDD_HHMMSS.log`

#### `frontend/start.sh` - 前端启动脚本
单独启动前端服务。

```bash
cd frontend
./start.sh
```

**功能：**
- ✅ 检查Node.js环境
- ✅ 自动安装依赖（如果未安装）
- ✅ 自动检测并停止占用端口3000的服务
- ✅ 启动前端服务
- ✅ 记录PID和日志

**日志位置：** `frontend/logs/frontend_YYYYMMDD_HHMMSS.log`

---

## 🚀 快速开始

### 首次使用

1. **给脚本添加执行权限（仅首次）：**
```bash
chmod +x start-all.sh stop-all.sh
chmod +x backend/start.sh frontend/start.sh
```

2. **启动所有服务：**
```bash
./start-all.sh
```

3. **访问系统：**
- 前端：http://localhost:3000
- 后端API：http://localhost:8080
- 健康检查：http://localhost:8080/health

4. **默认登录信息：**
- 账号：`admin`
- 密码：`admin123`
- ⚠️ **首次登录后请立即修改密码！**

---

## 📊 服务管理

### 查看服务状态

```bash
# 检查后端服务
lsof -i :8080

# 检查前端服务
lsof -i :3000
```

### 查看日志

```bash
# 后端日志（实时）
tail -f backend/logs/backend_*.log

# 前端日志（实时）
tail -f frontend/logs/frontend_*.log

# 查看最新日志
ls -lt backend/logs/  # 查看后端日志列表
ls -lt frontend/logs/  # 查看前端日志列表
```

### 健康检查

```bash
# 后端健康检查
curl http://localhost:8080/health

# 前端访问测试
curl http://localhost:3000
```

---

## 🔧 故障排除

### 端口被占用

如果启动失败提示端口被占用：

```bash
# 手动停止所有服务
./stop-all.sh

# 或者手动杀死进程
lsof -ti:8080 | xargs kill -9  # 后端
lsof -ti:3000 | xargs kill -9  # 前端
```

### 服务无法启动

1. **检查日志文件**
```bash
# 查看最新的后端日志
tail -50 backend/logs/backend_*.log | tail -50

# 查看最新的前端日志
tail -50 frontend/logs/frontend_*.log | tail -50
```

2. **检查环境**
```bash
# 检查Go版本
go version

# 检查Node.js版本
node --version
npm --version
```

3. **检查数据库连接**
```bash
# PostgreSQL
psql -U fnuser -d feiniu_user -c "SELECT 1"

# Redis
redis-cli ping
```

### 前端依赖问题

```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

---

## 📝 日志说明

### 后端日志内容

- 服务启动信息
- 数据库连接状态
- 超级管理员初始化
- 飞牛影视同步状态
- API请求日志
- 错误和警告信息

### 前端日志内容

- Vite服务器启动信息
- 编译状态
- 热更新信息
- 警告和错误

---

## 🎯 开发模式

### 后端开发

```bash
cd backend

# 直接运行（不使用脚本）
go run cmd/server/main.go

# 或使用脚本（推荐）
./start.sh
```

### 前端开发

```bash
cd frontend

# 直接运行（不使用脚本）
npm run dev

# 或使用脚本（推荐）
./start.sh
```

---

## ⚙️ 配置说明

### 后端配置

配置文件：`backend/config/config.yaml`

关键配置：
- 数据库连接
- Redis连接
- JWT密钥
- 飞牛影视API配置
- 邮件服务配置

### 前端配置

环境变量：`frontend/.env`

关键配置：
- API地址
- 端口号

---

## 🔐 安全提示

1. **首次登录后立即修改admin密码**
2. **不要在生产环境使用默认密码**
3. **定期检查日志文件，清理过期日志**
4. **保护好配置文件中的敏感信息**

---

## 📞 技术支持

如果遇到问题：

1. 查看日志文件
2. 检查服务状态
3. 验证配置文件
4. 查看错误提示

---

**版本：** v1.0  
**更新时间：** 2025-12-07  
**作者：** Cascade AI
