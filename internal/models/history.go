package models

import "time"

type Transaction struct {
	// UTC transaction time
	Time   time.Time `json:"time" db:"time"`
	From   string    `json:"from" db:"from_wallet"`
	To     string    `json:"to" db:"to_wallet"`
	Amount float32   `json:"amount" db:"amount"`
}
