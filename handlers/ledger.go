package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"udhaar-manager/middleware"
	"udhaar-manager/models"
)

type LedgerHandler struct {
	DB   *sql.DB
	Tmpl *template.Template
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

func (h *LedgerHandler) EditEntryPage(w http.ResponseWriter, r *http.Request) {
	entry, err := h.entryFromRequest(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.Tmpl.ExecuteTemplate(w, "edit_entry.html", entry)
}

func (h *LedgerHandler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	shopID := middleware.ShopIDFromContext(r)
	customerID := r.PathValue("id")
	entryID := r.PathValue("entryID")
	entryType := r.FormValue("entry_type")
	amount := r.FormValue("amount")
	note := r.FormValue("note")
	entryDate := r.FormValue("entry_date")

	if entryType != "credit" && entryType != "payment" {
		http.Error(w, "invalid entry type", http.StatusBadRequest)
		return
	}
	parsedAmount, err := strconv.ParseFloat(amount, 64)
	if err != nil || parsedAmount <= 0 {
		http.Error(w, "amount must be greater than zero", http.StatusBadRequest)
		return
	}
	if _, err := time.Parse("2006-01-02", entryDate); err != nil {
		http.Error(w, "invalid entry date", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		UPDATE ledger_entries
		SET entry_type = ?, amount = ?, note = ?, entry_date = ?, edited_at = NOW()
		WHERE id = ? AND customer_id = ? AND shop_id = ?`,
		entryType, parsedAmount, note, entryDate, entryID, customerID, shopID,
	)
	if err != nil {
		http.Error(w, "failed to update ledger entry", http.StatusInternalServerError)
		return
	}
	updated, err := result.RowsAffected()
	if err != nil || updated == 0 {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/customers/"+customerID, http.StatusSeeOther)
}

func (h *LedgerHandler) entryFromRequest(r *http.Request) (models.LedgerEntry, error) {
	var entry models.LedgerEntry
	err := h.DB.QueryRow(`
		SELECT id, customer_id, entry_type, amount, note, entry_date, edited_at
		FROM ledger_entries
		WHERE id = ? AND customer_id = ? AND shop_id = ?`,
		r.PathValue("entryID"), r.PathValue("id"), middleware.ShopIDFromContext(r),
	).Scan(&entry.ID, &entry.CustomerID, &entry.EntryType, &entry.Amount, &entry.Note, &entry.EntryDate, &entry.EditedAt)
	return entry, err
}
