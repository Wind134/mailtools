# mailtools

邮件发送工具，发送人显示为 **Ping**。

## 项目结构

```
mailtools/
├── cmd/send/main.go      # CLI 入口
├── internal/mail/        # 核心发送逻辑
├── go.mod
└── README.md
```

## 配置

创建 `~/.config/mailtools/config.yaml`：

```yaml
smtp_host: smtp.example.com
smtp_port: 587
username: your@email.com
password: your-password-or-app-password
from_name: Ping
```

### 常用邮箱配置

| 邮箱 | SMTP Host | 端口 |
|------|-----------|------|
| Gmail | smtp.gmail.com | 465 (SSL) |
| QQ邮箱 | smtp.qq.com | 587 |
| 163邮箱 | smtp.163.com | 587 |

> Gmail/QQ 邮箱需使用 App Password，而非登录密码。

## 安装

```bash
cd ~/mailtools
go build -o ~/go/bin/ping-mail ./cmd/send
```

## 使用

```bash
# 直接指定邮件内容
ping-mail send -t recipient@example.com -s "邮件主题" -b "邮件正文"

# 从文件读取正文
ping-mail send -t recipient@example.com -s "邮件主题" -f body.html

# 指定配置文件
ping-mail send -c /path/to/config.yaml -t recipient@example.com -s "邮件主题" -b "邮件正文"
```

## 构建

```bash
go mod tidy
go build -o ping-mail ./cmd/send
```
