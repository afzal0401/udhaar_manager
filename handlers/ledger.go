package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"udhaar-manager/middleware"
	"udhaar-manager/models"
	"udhaar-manager/notify"
)

type LedgerHandler struct {
	DB     *sql.DB
	Notify *notify.Sender
}

// AddEntry inserts a credit/payment row and sends a WhatsApp/SMS notification with the updated total.
func (h *LedgerHandler) AddEntry(w http.ResponseWriter, r *http.Request) {
	shopID := middleware.ShopIDFromContext(r)
	customerID := r.PathValue("id")

	entryType := r.FormValue("entry_type")
	amount := r.FormValue("amount")
	note := r.FormValue("note")
	entryDate := r.FormValue("entry_date")
	if entryDate == "" {
		entryDate = time.Now().Format("2006-01-02")
	}
	if entryType != "credit" && entryType != "payment" {
		http.Error(w, "invalid entry type", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec(`
		INSERT INTO ledger_entries (shop_id, customer_id, entry_type, amount, note, entry_date)
		VALUES (?, ?, ?, ?, ?, ?)`,
		shopID, customerID, entryType, amount, note, entryDate,
	)
	if err != nil {
		http.Error(w, "failed to add entry", http.StatusInternalServerError)
		return
	}

	h.notifyCustomer(shopID, customerID)
	http.Redirect(w, r, "/customers/"+customerID, http.StatusSeeOther)
}

func (h *LedgerHandler) notifyCustomer(shopID uint64, customerID string) {
	var shopName string
	h.DB.QueryRow("SELECT name FROM shops WHERE id = ?", shopID).Scan(&shopName)

	var c models.Customer
	err := h.DB.QueryRow(`
		SELECT name, phone, notify_channel, opening_balance FROM customers WHERE id = ?`, customerID,
	).Scan(&c.Name, &c.Phone, &c.NotifyChannel, &c.OpeningBalance)
	if err != nil || c.Phone == "" {
		return
	}

	var total float64
	h.DB.QueryRow(`
		SELECT COALESCE(SUM(CASE WHEN entry_type='credit' THEN amount ELSE -amount END),0)
		FROM ledger_entries WHERE customer_id = ?`, customerID,
	).Scan(&total)
	total += c.OpeningBalance

	message := notify.FormatReminderMessage(shopName, c.Name, total)
	h.Notify.Send(c.Phone, c.NotifyChannel, message)
}
