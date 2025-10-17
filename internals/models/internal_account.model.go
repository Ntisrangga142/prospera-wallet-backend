package models

import "time"

type InternalAccount struct {
	ID        int       `json:"id"`
	Name      *string   `json:"name"`
	Img       *string   `json:"img"`
	Tax       int       `json:"tax"`
	CreatedAt time.Time `json:"created_at"`
}
