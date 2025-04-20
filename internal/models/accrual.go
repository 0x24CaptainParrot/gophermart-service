package models

import (
	"encoding/json"
	"strconv"
)

type AccrualResponse struct {
	Order      int64   `json:"order,omitempty"`
	Status     string  `json:"status,omitempty"`
	Accrual    float64 `json:"accrual,omitempty"`
	StatusCode int     `json:"-"`
}

func (ar *AccrualResponse) UnmarshalJSON(data []byte) (err error) {
	type AccRespAlias AccrualResponse

	aliasVal := &struct {
		*AccRespAlias
		Order string `json:"order"`
	}{
		AccRespAlias: (*AccRespAlias)(ar),
	}

	if err = json.Unmarshal(data, aliasVal); err != nil {
		return
	}
	var num int
	num, err = strconv.Atoi(aliasVal.Order)
	ar.Order = int64(num)
	return
}
