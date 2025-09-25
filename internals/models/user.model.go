package models

type User struct {
	FullName    *string `json:"full_name"`
	PhoneNumber *string `json:"phone_number"`
	Avatar      *string `json:"avatar"`
}

type UserHistoryTransactions struct {
	ID           int           `json:"sender_id"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	TransactionID   int     `json:"transaction_id"`
	ReceiverID      int     `json:"receiver_id"`
	Avatar          *string `json:"avatar"`
	FullName        *string `json:"receiver_name"`
	PhoneNumber     *string `json:"receiver_phone"`
	TransactionType string  `json:"transaction_type"`
	Total           int     `json:"total"`
}
