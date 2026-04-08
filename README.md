# MailTools

邮件发送工具，支持 CLI 和 Web 界面，发送人显示为 **Ping**。

## 功能特性

- **CLI 工具** - 命令行发送邮件
- **Web 界面** - 浏览器发送邮件，支持附件
- **用户系统** - 注册、登录、管理员权限
- **附件支持** - 支持多文件上传，中文文件名处理

## 项目结构

```
mailtools/
├── cmd/
│   ├── send/            # CLI 入口 (ping-mail)
│   ├── web/              # Web 服务 (mailtools-web)
│   └── createuser/       # 用户创建工具
├── internal/
│   ├── mail/            # 邮件发送核心逻辑
│   ├── db/               # 数据库操作 (GORM)
│   ├── config/           # 配置文件加载 (Viper)
│   └── web/              # HTTP 处理器
├── frontend/             # Next.js 前端
├── scripts/              # 脚本和 systemd 配置
├── go.mod
└── README.md
```

## 快速开始

### 1. 配置邮件

创建 `~/.config/mailtools/config.toml`：

```toml
[database]
host = "localhost"
port = 3306
user = "root"
password = "your_db_password"
name = "mailtools"

[app]
host = "0.0.0.0"
port = 8080

[jwt]
secret = "change_this_to_random_string"
expire_hours = 24
```

### 2. 初始化数据库

```bash
mysql -u root -p < scripts/init.sql
```

### 3. 构建

```bash
cd ~/mailtools
go build -o ~/go/bin/mailtools-web ./cmd/web
go build -o ~/go/bin/ping-mail ./cmd/send
go build -o ~/go/bin/createuser ./cmd/createuser
```

### 4. 创建管理员账户

```bash
~/go/bin/createuser -admin admin your_password
```

### 5. 启动服务

**Web 服务 (systemd)**:
```bash
sudo cp scripts/mailtools-web.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable mailtools-web
sudo systemctl start mailtools-web
```

**前端开发**:
```bash
cd frontend
npm install
npm run dev
```

访问 http://localhost:3000

## CLI 使用

```bash
# 发送邮件
ping-mail send -t recipient@example.com -s "主题" -b "正文"

# 发送带附件的邮件
ping-mail send -t recipient@example.com -s "主题" -b "正文" -a file1.pdf -a file2.jpg
```

## 常用邮箱 SMTP 配置

| 邮箱 | SMTP Host | 端口 |
|------|----------|------|
| Gmail | smtp.gmail.com | 465 (SSL) |
| QQ邮箱 | smtp.qq.com | 587 |
| 163邮箱 | smtp.163.com | 587 |

> Gmail/QQ 邮箱需使用 App Password，而非登录密码。

## API 接口

| 接口 | 方法 | 描述 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/api/login` | POST | 登录 |
| `/api/register` | POST | 注册 |
| `/api/send` | POST | 发送邮件 |
| `/api/admin/users` | GET/POST | 用户管理 (管理员) |
| `/api/admin/users/:id` | PUT/DELETE | 用户编辑/删除 (管理员) |

## 技术栈

- **后端**: Go, Gin, GORM, JWT
- **前端**: Next.js, React, TailwindCSS
- **数据库**: MariaDB/MySQL
