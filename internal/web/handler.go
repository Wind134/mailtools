package web

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"mailtools/internal/db"
	"mailtools/internal/mail"
)

type SendMailRequest struct {
	To      string `form:"to" binding:"required"`
	Subject string `form:"subject" binding:"required"`
	Body    string `form:"body"`
}

type SendMailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UserListResponse struct {
	Users []UserResponse `json:"users"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6"`
	IsAdmin  bool   `json:"is_admin"`
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	IsAdmin  bool   `json:"is_admin"`
}

func InitRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(corsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().Format(time.RFC3339)})
	})

	r.POST("/api/login", HandleLogin)
	r.POST("/api/register", HandleRegister)

	auth := r.Group("/api")
	auth.Use(AuthMiddleware())
	{
		auth.POST("/send", HandleSendMail)
		auth.POST("/upload", HandleUpload)

		admin := auth.Group("/admin")
		admin.Use(AdminMiddleware())
		{
			admin.GET("/users", HandleListUsers)
			admin.POST("/users", HandleCreateUser)
			admin.PUT("/users/:id", HandleUpdateUser)
			admin.DELETE("/users/:id", HandleDeleteUser)
		}
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func HandleSendMail(c *gin.Context) {
	to := c.PostForm("to")
	subject := c.PostForm("subject")
	body := c.PostForm("body")

	if to == "" || subject == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to and subject are required"})
		return
	}

	if body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}

	form, err := c.MultipartForm()
	var attachments []string
	var tempFiles []string

	if err == nil && form != nil && form.File["attachments"] != nil {
		files := form.File["attachments"]
		for _, fileHeader := range files {
			filename := decodeFilename(fileHeader.Filename)
			filename = sanitizeFilename(filename)
			tmpDir := os.TempDir()
			tmpPath := filepath.Join(tmpDir, fmt.Sprintf("mailtools-%d-%s", time.Now().UnixNano(), filename))

			file, err := fileHeader.Open()
			if err != nil {
				continue
			}

			out, err := os.Create(tmpPath)
			if err != nil {
				file.Close()
				continue
			}

			_, err = io.Copy(out, file)
			file.Close()
			out.Close()

			if err != nil {
				os.Remove(tmpPath)
				continue
			}

			attachments = append(attachments, tmpPath)
			tempFiles = append(tempFiles, tmpPath)
		}
	}

	cfg, err := mail.LoadConfig(mail.BuildConfigPath())
	if err != nil {
		cleanupTempFiles(tempFiles)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("load config: %v", err)})
		return
	}

	sender := mail.NewSender(cfg)
	if err := sender.SendWithAttachments(to, subject, body, attachments); err != nil {
		cleanupTempFiles(tempFiles)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("send mail: %v", err)})
		return
	}

	cleanupTempFiles(tempFiles)

	c.JSON(http.StatusOK, SendMailResponse{
		Success: true,
		Message: fmt.Sprintf("Email sent to %s successfully", to),
	})
}

func cleanupTempFiles(paths []string) {
	for _, p := range paths {
		os.Remove(p)
	}
}

func HandleUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("get file: %v", err)})
		return
	}
	defer file.Close()

	filename := decodeFilename(header.Filename)
	filename = sanitizeFilename(filename)

	tmpDir := os.TempDir()
	tmpPath := filepath.Join(tmpDir, fmt.Sprintf("mailtools-%d-%s", time.Now().UnixNano(), filename))

	out, err := os.Create(tmpPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("create file: %v", err)})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		os.Remove(tmpPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("save file: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"path":     tmpPath,
		"filename": filename,
		"size":     header.Size,
	})
}

func HandleListUsers(c *gin.Context) {
	users, err := db.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get users"})
		return
	}

	response := make([]UserResponse, len(users))
	for i, u := range users {
		response[i] = UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			IsAdmin:   u.IsAdmin,
			CreatedAt: u.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, UserListResponse{Users: response})
}

func HandleCreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	existing, _ := db.GetUserByUsername(req.Username)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	hash, err := bcryptGenerateFromPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := db.CreateUser(req.Username, hash, req.IsAdmin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created successfully"})
}

func HandleUpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	existing, _ := db.GetUserByUsername(req.Username)
	if existing != nil && existing.ID != uint(id) {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	if err := db.UpdateUser(uint(id), req.Username, req.IsAdmin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}

func HandleDeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	userToDelete, _ := db.GetUserByID(uint(id))
	if userToDelete != nil && userToDelete.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete admin user"})
		return
	}

	if err := db.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

func bcryptGenerateFromPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func decodeFilename(filename string) string {
	if strings.Contains(filename, "*=") {
		parts := strings.Split(filename, "*'")
		if len(parts) == 3 {
			_, charset, encoded := parts[0], parts[1], parts[2]
			_ = charset
			if decoded, err := url.QueryUnescape(encoded); err == nil {
				return decoded
			}
		}
	}

	if strings.HasPrefix(filename, "=?UTF-8?B?") || strings.HasPrefix(filename, "=?GBK?B?") {
		start := strings.Index(filename, "?") + 1
		end := strings.LastIndex(filename, "?")
		if start > 0 && end > start {
			encoded := filename[start+2 : end]
			if decoded, err := base64.StdEncoding.DecodeString(encoded); err == nil {
				return string(decoded)
			}
		}
	}

	return filename
}

func sanitizeFilename(filename string) string {
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "_")
	filename = strings.ReplaceAll(filename, "\n", "")
	filename = strings.ReplaceAll(filename, "\r", "")
	filename = strings.ReplaceAll(filename, "\x00", "")

	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		name := filename[:255-len(ext)]
		filename = name + ext
	}

	return filename
}
