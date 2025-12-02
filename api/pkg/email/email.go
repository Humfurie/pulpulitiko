package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmailService struct {
	apiKey  string
	fromEmail string
	fromName  string
	siteURL   string
}

type SendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

type SendEmailResponse struct {
	Id string `json:"id"`
}

type ResendError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Name       string `json:"name"`
}

func NewEmailService(apiKey, fromEmail, fromName, siteURL string) *EmailService {
	return &EmailService{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
		siteURL:   siteURL,
	}
}

func (s *EmailService) Send(to, subject, html string) error {
	if s.apiKey == "" {
		return fmt.Errorf("email service not configured: missing API key")
	}

	payload := SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var resendErr ResendError
		if err := json.NewDecoder(resp.Body).Decode(&resendErr); err != nil {
			return fmt.Errorf("email send failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("email send failed: %s", resendErr.Message)
	}

	return nil
}

func (s *EmailService) SendPasswordReset(to, resetToken string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.siteURL, resetToken)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
        <h1 style="color: white; margin: 0; font-size: 24px;">Password Reset Request</h1>
    </div>
    <div style="background: #f9fafb; padding: 30px; border-radius: 0 0 10px 10px;">
        <p>Hi,</p>
        <p>We received a request to reset your password. Click the button below to create a new password:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: 600;">Reset Password</a>
        </div>
        <p style="color: #666; font-size: 14px;">This link will expire in 1 hour.</p>
        <p style="color: #666; font-size: 14px;">If you didn't request a password reset, you can safely ignore this email.</p>
        <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 30px 0;">
        <p style="color: #999; font-size: 12px; text-align: center;">
            If the button doesn't work, copy and paste this link into your browser:<br>
            <a href="%s" style="color: #667eea;">%s</a>
        </p>
    </div>
</body>
</html>
`, resetURL, resetURL, resetURL)

	return s.Send(to, "Reset your password", html)
}

// IsConfigured returns true if the email service has an API key configured
func (s *EmailService) IsConfigured() bool {
	return s.apiKey != ""
}
