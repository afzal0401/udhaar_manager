package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

type contextKey string

const ShopIDKey contextKey = "shopID"

// RequireAuth checks the session cookie against the sessions table and injects shop_id into the request context.
func RequireAuth(dbConn *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			var shopID uint64
			var expiresAt time.Time
			err = dbConn.QueryRow(
				"SELECT shop_id, expires_at FROM sessions WHERE token = ?", cookie.Value,
			).Scan(&shopID, &expiresAt)
			if err != nil || time.Now().After(expiresAt) {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), ShopIDKey, shopID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ShopIDFromContext(r *http.Request) uint64 {
	if v, ok := r.Context().Value(ShopIDKey).(uint64); ok {
		return v
	}
	return 0
}
