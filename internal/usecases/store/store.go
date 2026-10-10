package store

import "context"

type Settings struct {
	Name          string  `json:"store_name"`
	Description   string  `json:"description"`
	LineURL       string  `json:"line_oa_url"`
	Fee           float64 `json:"delivery_fee_thb"`
	FreeThreshold float64 `json:"free_delivery_threshold_thb"`
	Accepting     bool    `json:"accepting_orders"`
}
type Repository interface {
	Get(context.Context) (Settings, error)
	Save(context.Context, Settings) (Settings, error)
}
