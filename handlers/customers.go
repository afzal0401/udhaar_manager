package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

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
	name := r.FormValue("name")
	phone := r.FormValue("phone")
	channel := r.FormValue("notify_channel")
	if channel == "" {
		channel = "sms"
	}
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec(
		"INSERT INTO customers (shop_id, name, phone, notify_channel) VALUES (?, ?, ?, ?)",
		shopID, name, phone, channel,
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

	h.Tmpl.ExecuteTemplate(w, "customer.html", map[string]any{
		"Customer": c,
		"Entries":  entries,
	})
}
