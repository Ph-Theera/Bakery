package store

import (
	"bakery/internal/shared"
	"math"
)

type Settings struct {
	Name          string  `json:"store_name"`
	Description   string  `json:"description"`
	LineURL       string  `json:"line_oa_url"`
	Fee           float64 `json:"delivery_fee_thb"`
	FreeThreshold float64 `json:"free_delivery_threshold_thb"`
	Accepting     bool    `json:"accepting_orders"`
}

func Validate(s Settings) error {
	if s.Name == "" || len(s.Name) > 200 || len(s.Description) > 5000 || math.IsNaN(s.Fee) || math.IsInf(s.Fee, 0) || s.Fee < 0 || s.Fee > 1e6 || math.IsNaN(s.FreeThreshold) || math.IsInf(s.FreeThreshold, 0) || s.FreeThreshold < 0 || s.FreeThreshold > 1e9 {
		return shared.ErrInvalid
	}
	return nil
}
