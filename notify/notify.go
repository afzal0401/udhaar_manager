package notify

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

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

	if channel == "whatsapp" && s.cfg.GupshupAPIKey != "" {
		return s.sendWhatsApp(phone, message)
	}
	if s.cfg.MSG91AuthKey != "" {
		return s.sendSMS(phone, message)
	}

	// No provider configured yet — log so you can verify the flow works end-to-end before paying for a provider.
	log.Printf("[notify:console] to=%s channel=%s message=%q", phone, channel, message)
	return nil
}

func (s *Sender) sendWhatsApp(phone, message string) error {
	// TODO: replace with real Gupshup/AiSensy/Interakt template API call once approved.
	form := url.Values{}
	form.Set("phone", phone)
	form.Set("message", message)
	log.Printf("[notify:whatsapp:stub] to=%s message=%q", phone, message)
	_ = form
	return nil
}

func (s *Sender) sendSMS(phone, message string) error {
	// TODO: replace with real MSG91/Twilio SMS API call.
	req, err := http.NewRequest(http.MethodGet, "https://api.msg91.com/api/v5/otp", nil)
	if err != nil {
		return err
	}
	_ = req
	log.Printf("[notify:sms:stub] to=%s message=%q", phone, message)
	return nil
}

func FormatCreditMessage(shopName, customerName string, amount, total float64) string {
	return fmt.Sprintf("Hi %s, %s recorded ₹%.2f credit. Total due: ₹%.2f", customerName, shopName, amount, total)
}

func FormatReminderMessage(shopName, customerName string, total float64) string {
	return fmt.Sprintf("Reminder from %s: %s, your outstanding balance is ₹%.2f", shopName, customerName, total)
}
