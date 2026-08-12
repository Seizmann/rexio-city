package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/seizmann/rexio-city/backend/go/internal/config"
)

// EmailService sends transactional emails via Brevo (formerly Sendinblue).
// Only used for security-sensitive events — new-device login alerts.
type EmailService struct {
	apiKey    string
	fromEmail string
	fromName  string
}

func NewEmailService() *EmailService {
	cfg := config.Load()
	return &EmailService{
		apiKey:    cfg.BrevoAPIKey,
		fromEmail: cfg.BrevoFromEmail,
		fromName:  cfg.BrevoFromName,
	}
}

// SendNewDeviceAlert emails the user when their account is logged into from a new device/IP.
// This is a best-effort call — failure is logged but does not block login.
func (e *EmailService) SendNewDeviceAlert(toEmail, username, deviceInfo, ipAddress string, loginTime time.Time) error {
	if e.apiKey == "" {
		// Brevo not configured — skip silently in dev
		return nil
	}

	subject := "New sign-in to your RexiO City account"
	htmlContent := fmt.Sprintf(`
<p>Hi %s,</p>
<p>We detected a new sign-in to your RexiO City account.</p>
<table>
  <tr><td><strong>Time:</strong></td><td>%s (UTC)</td></tr>
  <tr><td><strong>Device:</strong></td><td>%s</td></tr>
  <tr><td><strong>IP Address:</strong></td><td>%s</td></tr>
</table>
<p>If this was you, no action is needed.</p>
<p>If you did not sign in, please <a href="https://city.rexio.pro/settings/sessions">revoke all sessions</a> immediately and change your password.</p>
<p>— The RexiO City Team</p>
`,
		username,
		loginTime.UTC().Format("2006-01-02 15:04:05"),
		deviceInfo,
		ipAddress,
	)

	payload := map[string]interface{}{
		"sender":  map[string]string{"name": e.fromName, "email": e.fromEmail},
		"to":      []map[string]string{{"email": toEmail, "name": username}},
		"subject": subject,
		"htmlContent": htmlContent,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("email marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("email request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", e.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("email send error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("brevo returned status %d", resp.StatusCode)
	}
	return nil
}
