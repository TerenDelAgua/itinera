package models

import "github.com/google/uuid"

type ExpenseCategory struct {
	Id       uuid.UUID `json:"id"`
	Slug     string    `json:"slug"`
	ColorHex string    `json:"color_hex"`
}
