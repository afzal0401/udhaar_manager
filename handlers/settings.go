package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strings"

	"udhaar-manager/middleware"
	"udhaar-manager/models"
)

type SettingsHandler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func (h *SettingsHandler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	shop, err := h.shopFromRequest(r)
	if err != nil {
		http.Error(w, "failed to load owner settings", http.StatusInternalServerError)
		return
	}

	h.Tmpl.ExecuteTemplate(w, "settings.html", shop)
}

func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	shopName := strings.TrimSpace(r.FormValue("shop_name"))
	if shopName == "" {
		http.Error(w, "shop name is required", http.StatusBadRequest)
		return
	}

	shopID := middleware.ShopIDFromContext(r)
	if _, err := h.DB.Exec("UPDATE shops SET name = ? WHERE id = ?", shopName, shopID); err != nil {
		http.Error(w, "failed to update shop name", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *SettingsHandler) shopFromRequest(r *http.Request) (models.Shop, error) {
	var shop models.Shop
	err := h.DB.QueryRow(`
		SELECT id, COALESCE(name, ''), COALESCE(owner_name, ''), owner_phone, plan
		FROM shops
		WHERE id = ?`, middleware.ShopIDFromContext(r)).Scan(
		&shop.ID, &shop.Name, &shop.OwnerName, &shop.OwnerPhone, &shop.Plan,
	)
	return shop, err
}
