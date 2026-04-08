package mail

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SMTPHost string `yaml:"smtp_host"`
	SMTPPort int    `yaml:"smtp_port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	FromName string `yaml:"from_name"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

type Sender struct {
	cfg *Config
}

func NewSender(cfg *Config) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) Send(to, subject, body string) error {
	return s.SendWithAttachments(to, subject, body, nil)
}

func (s *Sender) SendWithAttachments(to, subject, body string, attachments []string) error {
	from := s.cfg.Username
	name := s.cfg.FromName
	if name == "" {
		name = "Ping"
	}

	var msg []byte
	var err error
	if len(attachments) == 0 {
		msg = buildSimpleMessage(from, name, to, subject, body)
	} else {
		msg, err = buildMultipartMessage(from, name, to, subject, body, attachments)
		if err != nil {
			return fmt.Errorf("build multipart message: %w", err)
		}
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.SMTPHost)

	if s.cfg.SMTPPort == 465 {
		return s.sendViaImplicitTLS(addr, auth, from, to, msg)
	}
	return s.sendViaStartTLS(addr, auth, from, to, msg)
}

func (s *Sender) sendViaStartTLS(addr string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		conn.Close()
		return fmt.Errorf("new client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{ServerName: s.cfg.SMTPHost}
		if err = client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("write data: %w", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("close data: %w", err)
	}
	return client.Quit()
}

func (s *Sender) sendViaImplicitTLS(addr string, auth smtp.Auth, from, to string, msg []byte) error {
	tlsCfg := &tls.Config{ServerName: s.cfg.SMTPHost}
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	tlsConn := tls.Client(conn, tlsCfg)

	client, err := smtp.NewClient(tlsConn, s.cfg.SMTPHost)
	if err != nil {
		tlsConn.Close()
		return fmt.Errorf("new client: %w", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("write data: %w", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("close data: %w", err)
	}
	return client.Quit()
}

func encodeSubject(subject string) string {
	for _, r := range subject {
		if r > 127 || r == '\n' || r == '\r' || r == '=' {
			encoded := base64.StdEncoding.EncodeToString([]byte(subject))
			return fmt.Sprintf("=?UTF-8?B?%s?=", encoded)
		}
	}
	return subject
}

func buildSimpleMessage(from, fromName, to, subject, body string) []byte {
	header := fmt.Sprintf("From: %s <%s>\r\n", fromName, from)
	header += fmt.Sprintf("To: %s\r\n", to)
	header += fmt.Sprintf("Subject: %s\r\n", encodeSubject(subject))
	header += "MIME-Version: 1.0\r\n"
	header += "Content-Type: text/html; charset=UTF-8\r\n"
	header += "\r\n"
	return []byte(header + body)
}

func buildMultipartMessage(from, fromName, to, subject, body string, attachments []string) ([]byte, error) {
	boundary := "go-mail-boundary-" + randomString(16)

	header := fmt.Sprintf("From: %s <%s>\r\n", fromName, from)
	header += fmt.Sprintf("To: %s\r\n", to)
	header += fmt.Sprintf("Subject: %s\r\n", encodeSubject(subject))
	header += "MIME-Version: 1.0\r\n"
	header += fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary)
	header += "\r\n"

	var parts strings.Builder
	parts.WriteString("--" + boundary + "\r\n")
	parts.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	parts.WriteString("\r\n")
	parts.WriteString(body + "\r\n")

	for _, path := range attachments {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open attachment %s: %w", path, err)
		}
		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("read attachment %s: %w", path, err)
		}

		filename := escapeMimeFilename(filepath.Base(path))
		parts.WriteString("--" + boundary + "\r\n")
		parts.WriteString("Content-Type: application/octet-stream\r\n")
		parts.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
		parts.WriteString("Content-Transfer-Encoding: base64\r\n")
		parts.WriteString("\r\n")

		encoded := base64.StdEncoding.EncodeToString(data)
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			parts.WriteString(encoded[i:end] + "\r\n")
		}
		parts.WriteString("\r\n")
	}

	parts.WriteString("--" + boundary + "--\r\n")
	return []byte(header + parts.String()), nil
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	rand.Read(b)
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}

func escapeMimeFilename(filename string) string {
	filename = strings.ReplaceAll(filename, `"`, `""`)
	filename = strings.ReplaceAll(filename, "\n", " ")
	filename = strings.ReplaceAll(filename, "\r", " ")
	return filename
}

func BuildConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/", ".config", "mailtools", "config.yaml")
	}
	return filepath.Join(home, ".config", "mailtools", "config.yaml")
}

func EnsureConfigDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/"
	}
	dir := filepath.Join(home, ".config", "mailtools")
	return os.MkdirAll(dir, 0755)
}

func IsConfigExists() bool {
	_, err := os.Stat(BuildConfigPath())
	return err == nil
}
