package service

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/dmehra2102/budget-tracker/internal/config"
)

type EmailService interface {
	SendWelcomeEmail(ctx context.Context, email, firstName string) error
	SendPasswordResetEmail(ctx context.Context, email, firstName, token string) error
	SendPasswordChangedEmail(ctx context.Context, email, firstName string) error
	SendBudgetAlertEmail(ctx context.Context, email, firstName, budgetName string, percentage float64) error
}

type emailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) EmailService {
	return &emailService{cfg: cfg}
}

func (s *emailService) SendWelcomeEmail(ctx context.Context, email, firstName string) error {
	subject := "Welcome to Budget Tracker!"
	body := fmt.Sprintf(`
        <html>
            <body>
                <h2>Welcome %s!</h2>
                <p>Thank you for registering with Budget Tracker.</p>
                <p>Start managing your finances effectively today!</p>
            </body>
        </html>
    `, firstName)

	return s.sendEmail(email, subject, body)
}

func (s *emailService) SendPasswordResetEmail(ctx context.Context, email, firstName, token string) error {
	subject := "Password Reset Request"
	resetURL := fmt.Sprintf("https://yourapp.com/reset-password?token=%s", token)

	body := fmt.Sprintf(`
        <html>
            <body>
                <h2>Password Reset</h2>
                <p>Hi %s,</p>
                <p>You requested to reset your password. Click the link below:</p>
                <a href="%s">Reset Password</a>
                <p>This link expires in 1 hour.</p>
                <p>If you didn't request this, please ignore this email.</p>
            </body>
        </html>
    `, firstName, resetURL)

	return s.sendEmail(email, subject, body)
}

func (s *emailService) SendPasswordChangedEmail(ctx context.Context, email, firstName string) error {
	subject := "Password Changed Successfully"
	body := fmt.Sprintf(`
        <html>
            <body>
                <h2>Password Changed</h2>
                <p>Hi %s,</p>
                <p>Your password has been changed successfully.</p>
                <p>If you didn't make this change, please contact support immediately.</p>
            </body>
        </html>
    `, firstName)

	return s.sendEmail(email, subject, body)
}

func (s *emailService) SendBudgetAlertEmail(ctx context.Context, email, firstName, budgetName string, percentage float64) error {
	subject := fmt.Sprintf("Budget Alert: %s", budgetName)
	body := fmt.Sprintf(`
        <html>
            <body>
                <h2>Budget Alert</h2>
                <p>Hi %s,</p>
                <p>Your budget "%s" has reached %.0f%% of its limit.</p>
                <p>Consider reviewing your expenses to stay on track.</p>
            </body>
        </html>
    `, firstName, budgetName, percentage)

	return s.sendEmail(email, subject, body)
}

func (s *emailService) sendEmail(to, subject, htmlBody string) error {
	auth := smtp.PlainAuth("", s.cfg.Email.SMTPUsername, s.cfg.Email.SMTPPassword, s.cfg.Email.SMTPHost)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-version: 1.0;\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\";\r\n"+
		"\r\n"+
		"%s\r\n",
		s.cfg.Email.FromAddress, to, subject, htmlBody)

	addr := fmt.Sprintf("%s:%d", s.cfg.Email.SMTPHost, s.cfg.Email.SMTPPort)
	return smtp.SendMail(addr, auth, s.cfg.Email.FromAddress, []string{to}, []byte(msg))
}
