package models

import (
	"encoding/json"
	"strconv"
	"time"
)

type Order struct {
	ID        string    `json:"id,omitempty"`
	UserID    int       `json:"user_id"`
	Number    int64     `json:"number"`
	Status    string    `json:"status,omitempty"`
	Accrual   float64   `json:"accrual"`
	CreatedAt time.Time `json:"created_at"`
}

func (o Order) IsEmpty() bool {
	return o.Number == 0 && o.Status == "" && o.Accrual == 0
}

func (o Order) MarshalJSON() ([]byte, error) {
	type OrderAlias Order

	aliasVal := struct {
		OrderAlias
		Number string `json:"number"`
	}{
		OrderAlias: OrderAlias(o),
		Number:     strconv.Itoa(int(o.Number)),
	}

	return json.Marshal(aliasVal)
}
