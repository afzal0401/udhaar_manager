package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"udhaar-manager/middleware"
)

type LedgerHandler struct {
	DB *sql.DB
}

// AddEntry inserts a credit/payment row for the selected customer.
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

	http.Redirect(w, r, "/customers/"+customerID, http.StatusSeeOther)
}
