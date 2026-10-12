package orders

import (
	"bakery/internal/shared"
	"strings"
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

type PricedItem struct {
	ID, Name string
	Quantity int
	Price    int64
}

func ValidateQuote(c Command) error {
	if len(c.Items) == 0 || len(c.Items) > 100 || (c.Fulfillment.Type != "delivery" && c.Fulfillment.Type != "pickup") || c.Fulfillment.Type == "delivery" && (strings.TrimSpace(c.Fulfillment.Address) == "" || len(c.Fulfillment.Address) > 2000) || c.Fulfillment.Type == "pickup" && c.Fulfillment.PickupAt == nil {
		return shared.ErrInvalid
	}
	seen := map[string]bool{}
	for _, i := range c.Items {
		if !shared.UUID(i.ProductID) || i.Quantity < 1 || i.Quantity > 1000 || seen[i.ProductID] {
			return shared.ErrInvalid
		}
		seen[i.ProductID] = true
	}
	return nil
}
func ValidateCreate(c Command) error {
	if err := ValidateQuote(c); err != nil {
		return err
	}
	if strings.TrimSpace(c.Customer.Name) == "" || len([]rune(c.Customer.Name)) > 160 || strings.TrimSpace(c.Customer.Phone) == "" || len(c.Customer.Phone) > 40 || len(c.Key) < 16 || len(c.Key) > 100 || (c.PaymentMethod != "bank_transfer" && c.PaymentMethod != "promptpay") {
		return shared.ErrInvalid
	}
	return nil
}
func ValidateTransition(from, to string) error {
	next := map[string][]string{"pending_payment": {"cancelled"}, "pending_verification": {"cancelled"}, "confirmed": {"preparing"}, "preparing": {"ready_for_pickup", "shipped"}, "ready_for_pickup": {"completed"}, "shipped": {"completed"}}
	if from == to {
		return nil
	}
	for _, s := range next[from] {
		if s == to {
			return nil
		}
	}
	return shared.ErrConflict
}
func CanSubmit(order, payment string) bool {
	return (order == "pending_payment" || order == "pending_verification") && (payment == "pending" || payment == "rejected" || payment == "needs_resubmission" || payment == "submitted")
}
func ValidateReview(order, payment, target, reason string) error {
	if target != "approved" && target != "rejected" {
		return shared.ErrInvalid
	}
	if target == "rejected" && (strings.TrimSpace(reason) == "" || len(reason) > 2000) {
		return shared.ErrInvalid
	}
	if payment == target {
		return nil
	}
	if payment != "submitted" || order != "pending_verification" {
		return shared.ErrConflict
	}
	return nil
}
