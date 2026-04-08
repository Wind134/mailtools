package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mailtools/internal/config"
	"mailtools/internal/db"
	"mailtools/internal/web"
)

var flagConfig = flag.String("config", "", "config file path")

func main() {
	flag.Parse()

	configPath := *flagConfig
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("failed to get home dir")
		}
		configPath = fmt.Sprintf("%s/.config/mailtools/config.toml", home)
	}

	_, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := db.Init(config.GetDSN()); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	router := web.InitRouter()

	addr := fmt.Sprintf("%s:%d", config.AppCfg.App.Host, config.AppCfg.App.Port)
	log.Printf("Starting mailtools web server on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func init() {
	signal.Ignore(syscall.SIGPIPE)
}
