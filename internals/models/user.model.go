package models

import "time"

type Profile struct {
	FullName    *string `json:"full_name"`
	PhoneNumber *string `json:"phone_number"`
	Avatar      *string `json:"avatar"`
	Verified    bool    `json:"verified"`
}

type User struct {
	ID          int     `json:"user_id"`
	FullName    *string `json:"full_name"`
	PhoneNumber *string `json:"phone_number"`
	Avatar      *string `json:"avatar"`
	IsVerified  bool    `json:"is_verified"`
}

type TransactionHistory struct {
	ID                int       `db:"id" json:"id"`
	Type              string    `db:"type" json:"type"`                           // jenis transaksi (transfer, topup, withdraw, dll)
	Total             float64   `db:"total" json:"total"`                         // total nilai transaksi
	Direction         string    `db:"direction" json:"direction"`                 // debit atau credit
	CounterpartyType  string    `db:"counterparty_type" json:"counterparty_type"` // wallet / internal
	CounterpartyName  string    `db:"counterparty_name" json:"counterparty_name"`
	CounterpartyImg   *string   `db:"counterparty_img" json:"counterparty_img,omitempty"`
	CounterpartyPhone *string   `db:"counterparty_phone" json:"counterparty_phone,omitempty"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}
