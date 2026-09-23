package handlers

import "testing"

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
