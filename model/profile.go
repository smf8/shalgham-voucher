package model

type Profile struct {
	ID          int64   `json:"id"`
	PhoneNumber string  `json:"phone_number"`
	Balance     float64 `json:"balance"`
}
