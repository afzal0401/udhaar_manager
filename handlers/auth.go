package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strings"
	"time"

	"udhaar-manager/utils"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	h.Tmpl.ExecuteTemplate(w, "login.html", nil)
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	h.Tmpl.ExecuteTemplate(w, "register.html", nil)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	shopName := strings.TrimSpace(r.FormValue("shop_name"))
	ownerName := strings.TrimSpace(r.FormValue("owner_name"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")
	if shopName == "" || ownerName == "" || phone == "" || email == "" || !validPassword(password) {
		http.Error(w, "shop name, owner name, phone, email, and a password of at least 8 characters are required", http.StatusBadRequest)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to secure password", http.StatusInternalServerError)
		return
	}

	result, err := h.DB.Exec(`
		INSERT INTO shops (name, owner_name, owner_phone, owner_email, password_hash)
		VALUES (?, ?, ?, ?, ?)`, shopName, ownerName, phone, email, string(passwordHash))
	if err != nil {
		http.Error(w, "an account already exists with this phone number or email", http.StatusConflict)
		return
	}

	shopID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "failed to create account", http.StatusInternalServerError)
		return
	}
	h.createSession(w, r, uint64(shopID))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	phone := strings.TrimSpace(r.FormValue("phone"))
	password := r.FormValue("password")
	if phone == "" || password == "" {
		http.Error(w, "phone number and password are required", http.StatusBadRequest)
		return
	}

	var shopID uint64
	var passwordHash string
	err := h.DB.QueryRow(
		"SELECT id, COALESCE(password_hash, '') FROM shops WHERE owner_phone = ? AND is_active = TRUE", phone,
	).Scan(&shopID, &passwordHash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		http.Error(w, "invalid phone number or password", http.StatusUnauthorized)
		return
	}
	h.createSession(w, r, shopID)
}

func validPassword(password string) bool {
	return len(password) >= 8
}

func (h *AuthHandler) createSession(w http.ResponseWriter, r *http.Request, shopID uint64) {
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
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session_token"); err == nil {
		h.DB.Exec("DELETE FROM sessions WHERE token = ?", cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
