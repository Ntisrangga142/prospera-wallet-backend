package models

import "time"

type TransactionRequest struct {
	Type              string  `json:"type" binding:"required,oneof=top_up transfer"`
	Amount            int     `json:"amount" binding:"required,gt=0"`
	Total             int     `json:"total" binding:"required,gt=0"`
	Note              *string `json:"note,omitempty"`
	InternalAccountID *int    `json:"internal_account_id,omitempty"` // wajib kalau top_up
	ReceiverAccountID *int    `json:"receiver_account_id,omitempty"` // wajib kalau transfer
	PIN               string  `json:"pin"`                           // PIN tidak dibutuhkan untuk topup
}

// Transaction merepresentasikan hasil insert
type Transaction struct {
	ID         int       `json:"id"`
	Type       string    `json:"type"`
	Amount     int       `json:"amount"`
	Total      int       `json:"total"`
	Note       *string   `json:"note,omitempty"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	CreatedAt  time.Time `json:"created_at"`
}
