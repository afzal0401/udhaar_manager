package main

import (
	"html/template"
	"log"
	"net/http"

	"udhaar-manager/config"
	appdb "udhaar-manager/db"
	"udhaar-manager/handlers"
	"udhaar-manager/middleware"
)

func main() {
	cfg := config.Load()

	conn, err := appdb.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close()

	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	authH := &handlers.AuthHandler{DB: conn, Tmpl: tmpl}
	custH := &handlers.CustomerHandler{DB: conn, Tmpl: tmpl}
	ledgerH := &handlers.LedgerHandler{DB: conn}
	dashH := &handlers.DashboardHandler{DB: conn, Tmpl: tmpl}
	settingsH := &handlers.SettingsHandler{DB: conn, Tmpl: tmpl}

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /login", authH.LoginPage)
	mux.HandleFunc("POST /login", authH.Login)
	mux.HandleFunc("GET /register", authH.RegisterPage)
	mux.HandleFunc("POST /register", authH.Register)
	mux.HandleFunc("POST /logout", authH.Logout)

	// Protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("GET /dashboard", dashH.Dashboard)
	protected.HandleFunc("GET /settings", settingsH.SettingsPage)
	protected.HandleFunc("POST /settings", settingsH.UpdateSettings)
	protected.HandleFunc("GET /customers/new", custH.NewCustomerPage)
	protected.HandleFunc("POST /customers", custH.CreateCustomer)
	protected.HandleFunc("GET /customers/{id}", custH.CustomerDetail)
	protected.HandleFunc("POST /customers/{id}/entries", ledgerH.AddEntry)

	mux.Handle("/dashboard", middleware.RequireAuth(conn)(protected))
	mux.Handle("/settings", middleware.RequireAuth(conn)(protected))
	mux.Handle("POST /customers", middleware.RequireAuth(conn)(protected))
	mux.Handle("/customers/", middleware.RequireAuth(conn)(protected))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	log.Printf("Udhaar Manager running on http://localhost:%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
