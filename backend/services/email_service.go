package services

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/thomas-illiet/ai-bridge/config"
	"github.com/thomas-illiet/ai-bridge/models"
)

// sendEmail sends an HTML email. No-op when SMTP_HOST is not configured.
func sendEmail(cfg *config.Config, to []string, subject, html string) error {
	if cfg.SMTPHost == "" || len(to) == 0 {
		return nil
	}
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	var auth smtp.Auth
	if cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	}
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		cfg.SMTPFrom, strings.Join(to, ", "), subject, html,
	)
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

// htmlLayout wraps content in a clean responsive email shell.
func htmlLayout(title, preheader, body string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>%s</title>
</head>
<body style="margin:0;padding:0;background:#f1f5f9;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
<span style="display:none;max-height:0;overflow:hidden;">%s</span>
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f1f5f9;padding:40px 16px;">
  <tr><td align="center">
    <table width="100%%" cellpadding="0" cellspacing="0" style="max-width:560px;">

      <!-- Header -->
      <tr><td style="background:linear-gradient(135deg,#6366f1,#3b82f6);border-radius:12px 12px 0 0;padding:28px 32px;text-align:center;">
        <table cellpadding="0" cellspacing="0" style="margin:0 auto;">
          <tr>
            <td style="padding-right:10px;vertical-align:middle;">
              <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32">
                <rect width="32" height="32" rx="7" fill="rgba(255,255,255,0.15)"/>
                <path d="M5 21 Q16 7 27 21" fill="none" stroke="white" stroke-width="2.5" stroke-linecap="round"/>
                <line x1="9" y1="21" x2="9" y2="27" stroke="white" stroke-width="2.5" stroke-linecap="round"/>
                <line x1="23" y1="21" x2="23" y2="27" stroke="white" stroke-width="2.5" stroke-linecap="round"/>
                <line x1="5" y1="21" x2="27" y2="21" stroke="white" stroke-width="2" stroke-linecap="round" opacity="0.7"/>
                <circle cx="5" cy="21" r="2.2" fill="white"/>
                <circle cx="27" cy="21" r="2.2" fill="white"/>
              </svg>
            </td>
            <td style="vertical-align:middle;">
              <span style="color:white;font-size:20px;font-weight:700;letter-spacing:-0.3px;">AI Bridge</span>
            </td>
          </tr>
        </table>
      </td></tr>

      <!-- Body -->
      <tr><td style="background:white;padding:36px 32px;">
        %s
      </td></tr>

      <!-- Footer -->
      <tr><td style="background:#f8fafc;border-radius:0 0 12px 12px;padding:20px 32px;text-align:center;border-top:1px solid #e2e8f0;">
        <p style="margin:0;font-size:12px;color:#94a3b8;">
          This email was sent by AI Bridge. Please do not reply to this message.
        </p>
      </td></tr>

    </table>
  </td></tr>
</table>
</body>
</html>`, title, preheader, body)
}

func ctaButton(url, label, color string) string {
	return fmt.Sprintf(
		`<table cellpadding="0" cellspacing="0" width="100%%" style="margin:24px 0;">
  <tr><td align="center" style="background:%s;border-radius:8px;">
    <a href="%s" style="display:block;padding:14px 28px;color:white;font-size:15px;font-weight:600;text-decoration:none;border-radius:8px;text-align:center;">%s</a>
  </td></tr>
</table>`, color, url, label)
}

func infoBox(content string) string {
	return fmt.Sprintf(
		`<div style="background:#f8fafc;border-left:4px solid #6366f1;border-radius:0 8px 8px 0;padding:14px 18px;margin:20px 0;font-size:14px;color:#475569;line-height:1.6;">%s</div>`,
		content)
}

// SendNewRequestNotification notifies admin team of a new access request.
func SendNewRequestNotification(cfg *config.Config, user *models.RegisteredUser, req *models.AccessRequest) {
	to := adminRecipients(cfg)
	subject := fmt.Sprintf("[AI Bridge] New access request from %s", user.Username)

	adminURL := fmt.Sprintf("%s/admin", cfg.AppURL)
	body := fmt.Sprintf(`
<h2 style="margin:0 0 8px;font-size:22px;color:#1e293b;">New Access Request</h2>
<p style="margin:0 0 20px;color:#64748b;font-size:15px;">A user is requesting access to AI Bridge and needs your review.</p>

<table cellpadding="0" cellspacing="0" style="width:100%%;margin-bottom:20px;">
  <tr>
    <td style="padding:8px 0;border-bottom:1px solid #f1f5f9;width:120px;">
      <span style="font-size:13px;font-weight:600;color:#94a3b8;text-transform:uppercase;letter-spacing:0.05em;">Username</span>
    </td>
    <td style="padding:8px 0;border-bottom:1px solid #f1f5f9;">
      <span style="font-size:14px;color:#1e293b;font-weight:600;">%s</span>
    </td>
  </tr>
  <tr>
    <td style="padding:8px 0;border-bottom:1px solid #f1f5f9;">
      <span style="font-size:13px;font-weight:600;color:#94a3b8;text-transform:uppercase;letter-spacing:0.05em;">Email</span>
    </td>
    <td style="padding:8px 0;border-bottom:1px solid #f1f5f9;">
      <span style="font-size:14px;color:#475569;">%s</span>
    </td>
  </tr>
</table>

%s

%s`,
		user.Username,
		user.Email,
		infoBox(fmt.Sprintf("<strong>Reason:</strong><br/>%s", req.Reason)),
		ctaButton(adminURL, "Review in Admin Panel →", "#6366f1"),
	)

	_ = sendEmail(cfg, to, subject, htmlLayout("New Access Request — AI Bridge", "A new user is requesting access.", body))
}

// SendRequestApproved notifies the user that their request was approved.
func SendRequestApproved(cfg *config.Config, user *models.RegisteredUser) {
	if user.Email == "" {
		return
	}
	subject := "[AI Bridge] Your access request has been approved"
	appURL := cfg.AppURL
	body := fmt.Sprintf(`
<h2 style="margin:0 0 8px;font-size:22px;color:#1e293b;">Welcome aboard, %s! 🎉</h2>
<p style="margin:0 0 20px;color:#64748b;font-size:15px;">Great news — your access request has been <strong style="color:#059669;">approved</strong>. You can now log in and start using AI Bridge.</p>

%s

%s`,
		user.Username,
		infoBox("You now have access to all features available to your role. Head to the dashboard to get started."),
		ctaButton(appURL, "Go to AI Bridge →", "#059669"),
	)

	_ = sendEmail(cfg, []string{user.Email}, subject, htmlLayout("Access Approved — AI Bridge", "Your AI Bridge access request has been approved.", body))
}

// SendRequestRejected notifies the user that their request was rejected.
func SendRequestRejected(cfg *config.Config, user *models.RegisteredUser, note string) {
	if user.Email == "" {
		return
	}
	subject := "[AI Bridge] Your access request has been rejected"
	appURL := cfg.AppURL
	body := fmt.Sprintf(`
<h2 style="margin:0 0 8px;font-size:22px;color:#1e293b;">Access Request Not Approved</h2>
<p style="margin:0 0 20px;color:#64748b;font-size:15px;">Hi %s, your access request to AI Bridge has been reviewed and was not approved at this time.</p>

%s

<p style="font-size:14px;color:#64748b;margin:20px 0 0;">If you think this is an error or your situation has changed, you are welcome to submit a new request.</p>

%s`,
		user.Username,
		infoBox(fmt.Sprintf("<strong>Reason provided by the reviewer:</strong><br/>%s", note)),
		ctaButton(appURL, "Submit a New Request →", "#6366f1"),
	)

	_ = sendEmail(cfg, []string{user.Email}, subject, htmlLayout("Access Request Update — AI Bridge", "Your AI Bridge access request was not approved.", body))
}
