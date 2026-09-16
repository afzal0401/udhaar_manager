package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"time"

	"udhaar-manager/notify"
	"udhaar-manager/utils"
)

type AuthHandler struct {
	DB     *sql.DB
	Tmpl   *template.Template
	Notify *notify.Sender
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	h.Tmpl.ExecuteTemplate(w, "login.html", nil)
}

// RequestOTP creates the shop row if it doesn't exist yet, generates an OTP, and "sends" it via notify.
func (h *AuthHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	phone := r.FormValue("phone")
	if phone == "" {
		http.Error(w, "phone is required", http.StatusBadRequest)
		return
	}

	otp := utils.GenerateOTP()
	otpHash := utils.HashOTP(otp)
	expiresAt := time.Now().Add(5 * time.Minute)

	_, err := h.DB.Exec(`
		INSERT INTO shops (owner_phone, otp_hash, otp_expires_at)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE otp_hash = VALUES(otp_hash), otp_expires_at = VALUES(otp_expires_at)`,
		phone, otpHash, expiresAt,
	)
	if err != nil {
		http.Error(w, "failed to create otp", http.StatusInternalServerError)
		return
	}

	if err := h.Notify.SendOTP(phone, otp); err != nil {
		http.Error(w, "failed to send otp", http.StatusBadGateway)
		return
	}

	http.Redirect(w, r, "/login/verify?phone="+phone, http.StatusSeeOther)
}

func (h *AuthHandler) VerifyPage(w http.ResponseWriter, r *http.Request) {
	phone := r.URL.Query().Get("phone")
	h.Tmpl.ExecuteTemplate(w, "verify.html", map[string]string{"Phone": phone})
}

// VerifyOTP checks the submitted code and, on success, creates a session cookie.
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	phone := r.FormValue("phone")
	otp := r.FormValue("otp")

	var shopID uint64
	var shopName string
	var otpHash string
	var expiresAt time.Time
	err := h.DB.QueryRow(
		"SELECT id, COALESCE(name, ''), otp_hash, otp_expires_at FROM shops WHERE owner_phone = ?", phone,
	).Scan(&shopID, &shopName, &otpHash, &expiresAt)
	if err != nil {
		http.Error(w, "shop not found", http.StatusBadRequest)
		return
	}

	if time.Now().After(expiresAt) || utils.HashOTP(otp) != otpHash {
		http.Error(w, "invalid or expired OTP", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateSessionToken()
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	sessionExpiry := time.Now().Add(30 * 24 * time.Hour)
	_, err = h.DB.Exec(
		"INSERT INTO sessions (shop_id, token, expires_at) VALUES (?, ?, ?)",
		shopID, token, sessionExpiry,
	)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		Expires:  sessionExpiry,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	if shopName == "" || shopName == "My Shop" {
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session_token"); err == nil {
		h.DB.Exec("DELETE FROM sessions WHERE token = ?", cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
