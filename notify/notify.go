package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"udhaar-manager/config"
)

// Sender abstracts WhatsApp/SMS delivery so providers can be swapped without touching handlers.
type Sender struct {
	cfg *config.Config
}

func NewSender(cfg *config.Config) *Sender {
	return &Sender{cfg: cfg}
}

// Send tries WhatsApp first when configured and requested, then SMS, then falls back to console logging.
func (s *Sender) Send(phone, channel, message string) error {
	if phone == "" || channel == "none" {
		return nil
	}

	if channel == "whatsapp" && s.whatsAppConfigured() && s.cfg.WhatsAppMessageTemplate != "" {
		return s.sendWhatsAppTemplate(phone, s.cfg.WhatsAppMessageTemplate, message)
	}

	// No provider configured yet: log so local development can continue.
	log.Printf("[notify:console] to=%s channel=%s message=%q", phone, channel, message)
	return nil
}

// SendOTP delivers an OTP through WhatsApp when configured, otherwise logs it for local development.
func (s *Sender) SendOTP(phone, otp string) error {
	if phone == "" {
		return nil
	}
	if s.whatsAppConfigured() && s.cfg.WhatsAppOTPTemplate != "" {
		return s.sendWhatsAppTemplate(phone, s.cfg.WhatsAppOTPTemplate, otp)
	}
	log.Printf("[notify:console] to=%s channel=whatsapp message=\"Your Udhaar Manager OTP is: %s\"", phone, otp)
	return nil
}

func (s *Sender) whatsAppConfigured() bool {
	return s.cfg.WhatsAppAccessToken != "" &&
		s.cfg.WhatsAppPhoneNumberID != ""
}

func (s *Sender) sendWhatsAppTemplate(phone, templateName, bodyText string) error {
	if templateName == "" {
		return fmt.Errorf("WhatsApp template is not configured")
	}

	payload, err := json.Marshal(map[string]any{
		"messaging_product": "whatsapp",
		"to":                strings.TrimPrefix(phone, "+"),
		"type":              "template",
		"template": map[string]any{
			"name": templateName,
			"language": map[string]string{
				"code": s.cfg.WhatsAppTemplateLang,
			},
			"components": []map[string]any{
				{
					"type": "body",
					"parameters": []map[string]string{
						{"type": "text", "text": bodyText},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	endpoint := "https://graph.facebook.com/v21.0/" + s.cfg.WhatsAppPhoneNumberID + "/messages"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("authorization", "Bearer "+s.cfg.WhatsAppAccessToken)
	req.Header.Set("content-type", "application/json")

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return fmt.Errorf("send WhatsApp request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("WhatsApp rejected message: status %s: %s", resp.Status, body)
	}
	return nil
}

func FormatCreditMessage(shopName, customerName string, amount, total float64) string {
	return fmt.Sprintf("Hi %s, %s recorded ₹%.2f credit. Total due: ₹%.2f", customerName, shopName, amount, total)
}

func FormatReminderMessage(shopName, customerName string, total float64) string {
	return fmt.Sprintf("Reminder from %s: %s, your outstanding balance is ₹%.2f", shopName, customerName, total)
}
