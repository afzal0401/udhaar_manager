package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"net/mail"
	"strings"

	"udhaar-manager/middleware"
	"udhaar-manager/models"
)

type SettingsHandler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func normalizeShopProfile(shopName, ownerName, plan string) (string, string, string) {
	trimmedShop := strings.TrimSpace(shopName)
	if trimmedShop == "" {
		trimmedShop = "My Shop"
	}

	trimmedOwner := strings.TrimSpace(ownerName)
	trimmedPlan := strings.ToLower(strings.TrimSpace(plan))
	if trimmedPlan == "" || trimmedPlan == "unknown" {
		trimmedPlan = "trial"
	}
	if trimmedPlan != "trial" && trimmedPlan != "basic" && trimmedPlan != "pro" {
		trimmedPlan = "trial"
	}

	return trimmedShop, trimmedOwner, trimmedPlan
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
	shopName, ownerName, plan := normalizeShopProfile(
		r.FormValue("shop_name"),
		r.FormValue("owner_name"),
		r.FormValue("plan"),
	)
	ownerEmail := strings.ToLower(strings.TrimSpace(r.FormValue("owner_email")))
	if _, err := mail.ParseAddress(ownerEmail); err != nil {
		http.Error(w, "a valid recovery email is required", http.StatusBadRequest)
		return
	}

	shopID := middleware.ShopIDFromContext(r)
	if _, err := h.DB.Exec(
		"UPDATE shops SET name = ?, owner_name = ?, owner_email = ?, plan = ? WHERE id = ?",
		shopName, ownerName, ownerEmail, plan, shopID,
	); err != nil {
		http.Error(w, "failed to update shop settings", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *SettingsHandler) shopFromRequest(r *http.Request) (models.Shop, error) {
	var shop models.Shop
	err := h.DB.QueryRow(`
		SELECT id, COALESCE(name, ''), COALESCE(owner_name, ''), owner_phone, COALESCE(owner_email, ''), plan
		FROM shops
		WHERE id = ?`, middleware.ShopIDFromContext(r)).Scan(
		&shop.ID, &shop.Name, &shop.OwnerName, &shop.OwnerPhone, &shop.OwnerEmail, &shop.Plan,
	)
	if err == nil {
		shop.Name, shop.OwnerName, shop.Plan = normalizeShopProfile(shop.Name, shop.OwnerName, shop.Plan)
	}
	return shop, err
}
