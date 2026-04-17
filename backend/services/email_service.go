package services

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/thomas-illiet/ai-bridge/config"
	"github.com/thomas-illiet/ai-bridge/models"
)

func sendEmail(cfg *config.Config, to []string, subject, body string) error {
	if cfg.SMTPHost == "" || len(to) == 0 {
		return nil
	}
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	var auth smtp.Auth
	if cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	}
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		cfg.SMTPFrom, strings.Join(to, ", "), subject, body)
	return smtp.SendMail(addr, auth, cfg.SMTPFrom, to, []byte(msg))
}

func adminRecipients(cfg *config.Config) []string {
	var out []string
	for _, addr := range strings.Split(cfg.SMTPTo, ",") {
		if addr = strings.TrimSpace(addr); addr != "" {
			out = append(out, addr)
		}
	}
	return out
}

// SendNewRequestNotification notifies admin team of a new access request.
func SendNewRequestNotification(cfg *config.Config, user *models.RegisteredUser, req *models.AccessRequest) {
	to := adminRecipients(cfg)
	subject := fmt.Sprintf("[AI Bridge] New access request from %s", user.Username)
	body := fmt.Sprintf("User %s (%s) has submitted an access request.\n\nReason:\n%s\n\nReview it in the admin panel.",
		user.Username, user.Email, req.Reason)
	_ = sendEmail(cfg, to, subject, body)
}

// SendRequestApproved notifies the user that their request was approved.
func SendRequestApproved(cfg *config.Config, user *models.RegisteredUser) {
	if user.Email == "" {
		return
	}
	subject := "[AI Bridge] Your access request has been approved"
	body := fmt.Sprintf("Hi %s,\n\nYour access request has been approved. You can now log in and use AI Bridge.\n\nWelcome aboard!", user.Username)
	_ = sendEmail(cfg, []string{user.Email}, subject, body)
}

// SendRequestRejected notifies the user that their request was rejected.
func SendRequestRejected(cfg *config.Config, user *models.RegisteredUser, note string) {
	if user.Email == "" {
		return
	}
	subject := "[AI Bridge] Your access request has been rejected"
	body := fmt.Sprintf("Hi %s,\n\nYour access request has been reviewed and was not approved.\n\nReason: %s\n\nIf you believe this is an error, please submit a new request.", user.Username, note)
	_ = sendEmail(cfg, []string{user.Email}, subject, body)
}
