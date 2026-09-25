package models

import "time"

type Shop struct {
	ID         uint64
	Name       string
	OwnerName  string
	OwnerPhone string
	OwnerEmail string
	Plan       string
}

type Customer struct {
	ID             uint64
	ShopID         uint64
	Name           string
	Phone          string
	NotifyChannel  string
	OpeningBalance float64
	Outstanding    float64 // computed, not a DB column
	CreatedAt      time.Time
}

type LedgerEntry struct {
	ID               uint64
	ShopID           uint64
	CustomerID       uint64
	EntryType        string // "credit" or "payment"
	Amount           float64
	Note             string
	EntryDate        time.Time
	CreatedAt        time.Time
	EditedAt         *time.Time
	PreviousAmount   *float64 // set when the entry has been edited at least once
	BalanceAfter     float64  // computed while loading the customer ledger
	WhatsAppEntryURL string   // computed for credit entries with a customer phone number
	WhatsAppEditURL  string   // computed for edited entries with a customer phone number
}
