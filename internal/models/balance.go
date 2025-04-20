package models

import (
	"encoding/json"
	"strconv"
	"time"
)

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order int64   `json:"order"`
	Sum   float64 `json:"sum"`
}

func (w *WithdrawRequest) UnmarshalJSON(data []byte) (err error) {
	type WithdrawAlias WithdrawRequest

	withdrawVal := &struct {
		*WithdrawAlias
		Order string `json:"order"`
	}{
		WithdrawAlias: (*WithdrawAlias)(w),
	}

	if err = json.Unmarshal(data, withdrawVal); err != nil {
		return
	}

	var num int
	num, err = strconv.Atoi(withdrawVal.Order)
	w.Order = int64(num)
	return
}

type Withdrawal struct {
	Order       int64     `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (t Withdrawal) MarshalJSON() ([]byte, error) {
	type TransactionAlias Withdrawal

	trVal := struct {
		TransactionAlias
		Order string `json:"order"`
	}{
		TransactionAlias: TransactionAlias(t),
		Order:            strconv.Itoa(int(t.Order)),
	}

	return json.Marshal(trVal)
}
