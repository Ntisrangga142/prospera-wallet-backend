package models

import "time"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ChangePassword struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type BlacklistToken struct {
	Token     string        `json:"token"`
	ExpiresIn time.Duration `json:"expires_in"`
}

type PINRequest struct {
	PIN string `json:"pin" example:"123456"`
}

type ForgotRequest struct {
	Email string `json:"email"`
	Type  string `json:"type"`
}

type ForgotPasswordScan struct {
	ID        int       `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type PINResetRequest struct {
	PIN   string `json:"pin" example:"123456"`
	Token string `json:"token" example:"1d9a25ef-5a08-46f3-9c21-6d3a2e9e6f7a"`
}

type PasswordResetRequest struct {
	Password string `json:"password" example:"User!23456789"`
	Token    string `json:"token" example:"1d9a25ef-5a08-46f3-9c21-6d3a2e9e6f7a"`
}
