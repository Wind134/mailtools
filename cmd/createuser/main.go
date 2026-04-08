package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"

	"mailtools/internal/config"
	"mailtools/internal/db"
)

func main() {
	isAdmin := flag.Bool("admin", false, "create as admin user")
	flag.Parse()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		home, _ := os.UserHomeDir()
		configPath = fmt.Sprintf("%s/.config/mailtools/config.toml", home)
	}

	_, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := db.Init(config.GetDSN()); err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Usage: createuser [-admin] <username> <password>")
		os.Exit(1)
	}

	username := args[0]
	password := args[1]

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Failed to hash password: %v\n", err)
		os.Exit(1)
	}

	user := &db.User{
		Username:     username,
		PasswordHash: string(hash),
		IsAdmin:      *isAdmin,
	}

	result := db.DB.Create(user)
	if result.Error != nil {
		fmt.Printf("Failed to create user: %v\n", result.Error)
		os.Exit(1)
	}

	role := "user"
	if *isAdmin {
		role = "admin"
	}
	fmt.Printf("User '%s' created successfully as %s!\n", username, role)
}
