package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

	"udhaar-manager/middleware"
	"udhaar-manager/models"
)

type DashboardHandler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	shopID := middleware.ShopIDFromContext(r)

	rows, err := h.DB.Query(`
		SELECT c.id, c.name, c.phone, c.opening_balance,
		       c.opening_balance + COALESCE(SUM(CASE WHEN l.entry_type='credit' THEN l.amount ELSE -l.amount END), 0) AS outstanding
		FROM customers c
		LEFT JOIN ledger_entries l ON l.customer_id = c.id
		WHERE c.shop_id = ? AND c.is_active = TRUE
		GROUP BY c.id
		ORDER BY c.name ASC`, shopID)
	if err != nil {
		http.Error(w, "failed to load dashboard", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var customers []models.Customer
	var totalOutstanding float64
	for rows.Next() {
		var c models.Customer
		if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.OpeningBalance, &c.Outstanding); err != nil {
			continue
		}
		customers = append(customers, c)
		totalOutstanding += c.Outstanding
	}

	h.Tmpl.ExecuteTemplate(w, "dashboard.html", map[string]any{
		"Customers":        customers,
		"TotalOutstanding": totalOutstanding,
	})
}
