package handlers

import (
	"testing"
	"time"
)

func TestWhatsAppReminderURL(t *testing.T) {
	url := whatsAppReminderURL("+91 98123-45678", "Afzal Store", "Ravi", 500)
	want := "https://wa.me/919812345678?text=Hello+Ravi%2C+this+is+a+reminder+from+Afzal+Store.+Your+outstanding+balance+is+Rs.+500.00."
	if url != want {
		t.Fatalf("unexpected WhatsApp URL: %q", url)
	}
}

func TestWhatsAppReminderURLEmptyPhone(t *testing.T) {
	if url := whatsAppReminderURL("", "Afzal Store", "Ravi", 500); url != "" {
		t.Fatalf("expected no URL, got %q", url)
	}
}

func TestWhatsAppEntryURL(t *testing.T) {
	entryDate := time.Date(2026, time.September, 23, 0, 0, 0, 0, time.UTC)
	url := whatsAppEntryURL("+91 98123-45678", "Afzal Store", "Srikanth", entryDate, 1000, "rice and oil", 1500)
	want := "https://wa.me/919812345678?text=Hello+Srikanth%2C+on+23+Sep+2026+you+took+udhaar+of+Rs.+1000.00+for+rice+and+oil+from+Afzal+Store.+Your+total+amount+due+is+Rs.+1500.00."
	if url != want {
		t.Fatalf("unexpected entry WhatsApp URL: %q", url)
	}
}
