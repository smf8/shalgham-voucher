package model

type Transaction struct {
	ID        int     `json:"id"`
	ProfileID int     `json:"profile_id"`
	Amount    float64 `json:"amount"`
}
