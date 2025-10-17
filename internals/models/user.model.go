package models

import "time"

type Profile struct {
	FullName    *string `json:"full_name"`
	PhoneNumber *string `json:"phone_number"`
	Avatar      *string `json:"avatar"`
	Verified    bool    `json:"verified"`
	Email       string  `json:"email"`
}

type User struct {
	ID          int     `json:"id"`
	FullName    *string `json:"full_name"`
	PhoneNumber *string `json:"phone_number"`
	Avatar      *string `json:"avatar"`
	Verified    bool    `json:"verified"`
}

type TransactionHistory struct {
	ID                *int       `db:"id" json:"id"`
	Type              *string    `db:"type" json:"type"`
	Total             *float64   `db:"total" json:"total"`
	Direction         *string    `db:"direction" json:"direction"`
	CounterpartyType  *string    `db:"counterparty_type" json:"counterparty_type"`
	CounterpartyName  *string    `db:"counterparty_name" json:"counterparty_name"`
	CounterpartyImg   *string    `db:"counterparty_img" json:"counterparty_img,omitempty"`
	CounterpartyPhone *string    `db:"counterparty_phone" json:"counterparty_phone,omitempty"`
	CreatedAt         *time.Time `db:"created_at" json:"created_at"`
}

type DailySummary struct {
	Date         time.Time `json:"date"`
	TotalExpense int       `json:"total_expense"`
	TotalIncome  int       `json:"total_income"`
}

type WeeklySummary struct {
	WeekStart    time.Time `json:"week_start"`
	WeekEnd      time.Time `json:"week_end"`
	TotalExpense int       `json:"total_expense"`
	TotalIncome  int       `json:"total_income"`
}
