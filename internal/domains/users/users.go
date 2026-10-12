package users

type Customer struct {
	ID         string `json:"id"`
	Name       string `json:"display_name"`
	Phone      string `json:"phone"`
	OrderCount int    `json:"order_count"`
}
type Profile struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}
