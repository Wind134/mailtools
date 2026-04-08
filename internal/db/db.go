package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

type User struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Username     string `gorm:"uniqueIndex;size:64;not null" json:"username"`
	PasswordHash string `gorm:"size:255;not null" json:"-"`
	IsAdmin      bool   `gorm:"default:false" json:"is_admin"`
	CreatedAt    int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    int64  `gorm:"autoUpdateTime" json:"updated_at"`
}

func Init(dsn string) error {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("get underlying db: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	if err := DB.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	return nil
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	result := DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &user, nil
}

func GetUserByID(id uint) (*User, error) {
	var user User
	result := DB.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &user, nil
}

func CreateUser(username, passwordHash string, isAdmin bool) error {
	user := User{
		Username:     username,
		PasswordHash: passwordHash,
		IsAdmin:      isAdmin,
	}
	return DB.Create(&user).Error
}

func GetAllUsers() ([]User, error) {
	var users []User
	result := DB.Find(&users)
	return users, result.Error
}

func DeleteUser(id uint) error {
	return DB.Delete(&User{}, id).Error
}

func UpdateUser(id uint, username string, isAdmin bool) error {
	return DB.Model(&User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"username": username,
		"is_admin": isAdmin,
	}).Error
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
