package main

import (
	"fmt"
	"os"
	"strings"

	"mailtools/internal/mail"

	"github.com/spf13/cobra"
)

var (
	flagTo          string
	flagSubject     string
	flagBody        string
	flagFile        string
	flagConfig      string
	flagAttachments []string
)

var rootCmd = &cobra.Command{
	Use:   "send",
	Short: "Send email as Ping",
	RunE:  runSend,
}

func init() {
	rootCmd.Flags().StringVarP(&flagTo, "to", "t", "", "recipient email address (required)")
	rootCmd.Flags().StringVarP(&flagSubject, "subject", "s", "", "email subject (required)")
	rootCmd.Flags().StringVarP(&flagBody, "body", "b", "", "email body")
	rootCmd.Flags().StringVarP(&flagFile, "file", "f", "", "read body from file")
	rootCmd.Flags().StringSliceVarP(&flagAttachments, "attachment", "a", nil, "attach file (can be used multiple times)")
	rootCmd.Flags().StringVarP(&flagConfig, "config", "c", "", "config file path (default ~/.config/mailtools/config.yaml)")
	rootCmd.MarkFlagRequired("to")
	rootCmd.MarkFlagRequired("subject")
}

func runSend(cmd *cobra.Command, args []string) error {
	if flagBody == "" && flagFile == "" {
		return fmt.Errorf("body or file flag is required")
	}

	if flagBody != "" && flagFile != "" {
		fmt.Println("Warning: both body and file specified, using body")
	}

	if !isValidEmail(flagTo) {
		return fmt.Errorf("invalid email address: %s", flagTo)
	}

	cfgPath := flagConfig
	if cfgPath == "" {
		cfgPath = mail.BuildConfigPath()
	}

	cfg, err := mail.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	sender := mail.NewSender(cfg)

	body := flagBody
	if flagFile != "" {
		data, err := os.ReadFile(flagFile)
		if err != nil {
			return fmt.Errorf("read body file: %w", err)
		}
		body = string(data)
	}

	var attachments []string
	if len(flagAttachments) > 0 {
		attachments = flagAttachments
	}

	if err := sender.SendWithAttachments(flagTo, flagSubject, body, attachments); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	if len(attachments) > 0 {
		fmt.Printf("Email with %d attachment(s) sent to %s successfully.\n", len(attachments), flagTo)
	} else {
		fmt.Printf("Email sent to %s successfully.\n", flagTo)
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
