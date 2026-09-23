package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

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
		SELECT id, entry_type, amount, note, entry_date
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
		if err := rows.Scan(&e.ID, &e.EntryType, &e.Amount, &e.Note, &e.EntryDate); err != nil {
			continue
		}
		if e.EntryType == "credit" {
			balance += e.Amount
		} else {
			balance -= e.Amount
		}
		entries = append(entries, e)
	}
	c.Outstanding = balance
	var shopName string
	if err := h.DB.QueryRow("SELECT name FROM shops WHERE id = ?", shopID).Scan(&shopName); err != nil {
		http.Error(w, "failed to load shop", http.StatusInternalServerError)
		return
	}

	h.Tmpl.ExecuteTemplate(w, "customer.html", map[string]any{
		"Customer":            c,
		"Entries":             entries,
		"WhatsAppReminderURL": whatsAppReminderURL(c.Phone, shopName, c.Name, c.Outstanding),
	})
}

func whatsAppReminderURL(phone, shopName, customerName string, outstanding float64) string {
	var digits strings.Builder
	for _, character := range phone {
		if character >= '0' && character <= '9' {
			digits.WriteRune(character)
		}
	}
	if digits.Len() == 0 {
		return ""
	}
	message := fmt.Sprintf("Hello %s, this is a reminder from %s. Your outstanding balance is Rs. %.2f.", customerName, shopName, outstanding)
	return "https://wa.me/" + digits.String() + "?text=" + url.QueryEscape(message)
}
