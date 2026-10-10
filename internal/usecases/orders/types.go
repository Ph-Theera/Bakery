package orders

import (
	"context"
	"time"
)

type Item struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}
type Command struct {
	Customer struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	} `json:"customer"`
	Fulfillment struct {
		Type     string     `json:"type"`
		Address  string     `json:"address"`
		PickupAt *time.Time `json:"pickupAt"`
	} `json:"fulfillment"`
	Items         []Item `json:"items"`
	PaymentMethod string `json:"paymentMethod"`
	Key           string `json:"-"`
	UserID        string `json:"-"`
}
type Quote struct {
	Subtotal    float64 `json:"subtotal_thb"`
	DeliveryFee float64 `json:"delivery_fee_thb"`
	Total       float64 `json:"total_thb"`
}
type Line struct {
	Name     string  `json:"product_name_snapshot"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"unit_price_thb"`
}
type Data struct {
	ID            string    `json:"id"`
	Number        string    `json:"order_number"`
	PaymentID     string    `json:"payment_id"`
	Customer      string    `json:"customer_name_snapshot"`
	Total         float64   `json:"total_thb"`
	Status        string    `json:"order_status"`
	PaymentStatus string    `json:"payment_status"`
	CreatedAt     time.Time `json:"created_at"`
	Items         []Line    `json:"items"`
	AccessToken   string    `json:"access_token,omitempty"`
}
type Evidence struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
type Repository interface {
	Quote(context.Context, Command) (Quote, error)
	Create(context.Context, Command) (Data, error)
	Get(context.Context, string) (Data, error)
	List(context.Context, string, int, int) ([]Data, error)
	Status(context.Context, string, string, string, string) (Data, error)
	Note(context.Context, string, string, string) (Data, error)
	Slip(context.Context, string, string, string, []byte) (Evidence, error)
	Submit(context.Context, string, string) error
	Review(context.Context, string, string, string, string) (Data, error)
	Dashboard(context.Context) (map[string]any, error)
	PaymentOrder(context.Context, string) (string, error)
	Evidence(context.Context, string) ([]byte, string, error)
	Mine(context.Context, string, int, int) ([]Data, error)
	Owned(context.Context, string, string) (bool, error)
}
