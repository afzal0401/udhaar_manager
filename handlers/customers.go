package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"udhaar-manager/middleware"
	"udhaar-manager/models"
)

type CustomerHandler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func (h *CustomerHandler) NewCustomerPage(w http.ResponseWriter, r *http.Request) {
	h.Tmpl.ExecuteTemplate(w, "add_customer.html", nil)
}

func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	shopID := middleware.ShopIDFromContext(r)
	name := strings.TrimSpace(r.FormValue("name"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if phone != "" && whatsAppPhoneDigits(phone) == "" {
		http.Error(w, "phone must be a valid 10-digit mobile number", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec(
		"INSERT INTO customers (shop_id, name, phone, notify_channel) VALUES (?, ?, ?, 'none')",
		shopID, name, phone,
	)
	if err != nil {
		http.Error(w, "failed to add customer", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *CustomerHandler) CustomerDetail(w http.ResponseWriter, r *http.Request) {
	shopID := middleware.ShopIDFromContext(r)
	id := r.PathValue("id")

	var c models.Customer
	err := h.DB.QueryRow(`
		SELECT id, name, phone, notify_channel, opening_balance
		FROM customers WHERE id = ? AND shop_id = ?`, id, shopID,
	).Scan(&c.ID, &c.Name, &c.Phone, &c.NotifyChannel, &c.OpeningBalance)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	rows, err := h.DB.Query(`
		SELECT id, entry_type, amount, note, entry_date, edited_at, previous_amount
		FROM ledger_entries WHERE customer_id = ? ORDER BY entry_date ASC, id ASC`, c.ID)
	if err != nil {
		http.Error(w, "failed to load ledger", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entries []models.LedgerEntry
	balance := c.OpeningBalance
	for rows.Next() {
		var e models.LedgerEntry
		if err := rows.Scan(&e.ID, &e.EntryType, &e.Amount, &e.Note, &e.EntryDate, &e.EditedAt, &e.PreviousAmount); err != nil {
			continue
		}
		if e.EntryType == "credit" {
			balance += e.Amount
		} else {
			balance -= e.Amount
		}
		e.BalanceAfter = balance
		entries = append(entries, e)
	}
	c.Outstanding = balance
	var shopName string
	if err := h.DB.QueryRow("SELECT name FROM shops WHERE id = ?", shopID).Scan(&shopName); err != nil {
		http.Error(w, "failed to load shop", http.StatusInternalServerError)
		return
	}
	for index := range entries {
		entry := &entries[index]
		if entry.EntryType == "credit" {
			entry.WhatsAppEntryURL = whatsAppEntryURL(c.Phone, shopName, c.Name, entry.EntryDate, entry.Amount, entry.Note, entry.BalanceAfter)
		}
		if entry.EditedAt != nil && entry.PreviousAmount != nil {
			entry.WhatsAppEditURL = whatsAppEditURL(c.Phone, shopName, c.Name, entry.EntryDate, *entry.PreviousAmount, entry.Amount, entry.BalanceAfter)
		}
	}

	h.Tmpl.ExecuteTemplate(w, "customer.html", map[string]any{
		"Customer":            c,
		"Entries":             entries,
		"WhatsAppReminderURL": whatsAppReminderURL(c.Phone, shopName, c.Name, c.Outstanding),
	})
}

func whatsAppReminderURL(phone, shopName, customerName string, outstanding float64) string {
	phoneDigits := whatsAppPhoneDigits(phone)
	if phoneDigits == "" {
		return ""
	}
	message := fmt.Sprintf("Hello %s, this is a reminder from %s. Your outstanding balance is Rs. %.2f.", customerName, shopName, outstanding)
	return whatsAppURL(phoneDigits, message)
}

// isValidIndianMobileDigits reports whether digits is a usable WhatsApp number:
// either a bare 10-digit mobile number (6-9 leading digit) or one already prefixed with the 91 country code.
func isValidIndianMobileDigits(digits string) bool {
	switch len(digits) {
	case 10:
		return digits[0] >= '6' && digits[0] <= '9'
	case 12:
		return strings.HasPrefix(digits, "91") && digits[2] >= '6' && digits[2] <= '9'
	default:
		return false
	}
}

func whatsAppEntryURL(phone, shopName, customerName string, entryDate time.Time, amount float64, note string, balanceAfter float64) string {
	phoneDigits := whatsAppPhoneDigits(phone)
	if phoneDigits == "" {
		return ""
	}
	itemText := strings.TrimSpace(note)
	if itemText == "" {
		itemText = "items"
	}
	message := fmt.Sprintf(
		"Hello %s, on %s you took udhaar of Rs. %.2f for %s from %s. Your total amount due is Rs. %.2f.",
		customerName, entryDate.Format("02 Jan 2006"), amount, itemText, shopName, balanceAfter,
	)
	return whatsAppURL(phoneDigits, message)
}

func whatsAppEditURL(phone, shopName, customerName string, entryDate time.Time, previousAmount, newAmount, balanceAfter float64) string {
	phoneDigits := whatsAppPhoneDigits(phone)
	if phoneDigits == "" {
		return ""
	}
	message := fmt.Sprintf(
		"Hello %s, %s corrected your entry dated %s. The amount has been updated from Rs. %.2f to Rs. %.2f. Your total amount due is now Rs. %.2f.",
		customerName, shopName, entryDate.Format("02 Jan 2006"), previousAmount, newAmount, balanceAfter,
	)
	return whatsAppURL(phoneDigits, message)
}

// whatsAppPhoneDigits strips a phone number down to its digits and normalizes it to
// the 12-digit "91XXXXXXXXXX" form wa.me requires, returning "" if it isn't a valid Indian mobile number.
func whatsAppPhoneDigits(phone string) string {
	var digits strings.Builder
	for _, character := range phone {
		if character >= '0' && character <= '9' {
			digits.WriteRune(character)
		}
	}
	raw := digits.String()
	if !isValidIndianMobileDigits(raw) {
		return ""
	}
	if len(raw) == 10 {
		return "91" + raw
	}
	return raw
}

func whatsAppURL(phoneDigits, message string) string {
	if phoneDigits == "" {
		return ""
	}
	return "https://wa.me/" + phoneDigits + "?text=" + url.QueryEscape(message)
}
