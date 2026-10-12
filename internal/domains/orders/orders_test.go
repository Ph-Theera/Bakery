package orders

import (
	"bakery/internal/shared"
	"errors"
	"testing"
)

func TestStatusPolicy(t *testing.T) {
	for _, tc := range []struct {
		from, to string
		valid    bool
	}{{"pending_payment", "cancelled", true}, {"confirmed", "preparing", true}, {"preparing", "shipped", true}, {"preparing", "ready_for_pickup", true}, {"shipped", "completed", true}, {"confirmed", "completed", false}, {"confirmed", "cancelled", false}, {"pending_payment", "confirmed", false}} {
		e := ValidateTransition(tc.from, tc.to)
		if tc.valid && e != nil || !tc.valid && !errors.Is(e, shared.ErrConflict) {
			t.Fatalf("%s -> %s: %v", tc.from, tc.to, e)
		}
	}
}
func TestPaymentReviewPolicy(t *testing.T) {
	if e := ValidateReview("pending_verification", "submitted", "approved", ""); e != nil {
		t.Fatal(e)
	}
	if e := ValidateReview("pending_payment", "pending", "approved", ""); !errors.Is(e, shared.ErrConflict) {
		t.Fatal(e)
	}
	if e := ValidateReview("pending_verification", "submitted", "rejected", " "); !errors.Is(e, shared.ErrInvalid) {
		t.Fatal(e)
	}
	if e := ValidateReview("confirmed", "approved", "approved", ""); e != nil {
		t.Fatal(e)
	}
	if CanSubmit("cancelled", "pending") || CanSubmit("confirmed", "approved") {
		t.Fatal("invalid submission permitted")
	}
}
func TestCheckoutDuplicateValidation(t *testing.T) {
	var c Command
	c.Fulfillment.Type = "delivery"
	c.Fulfillment.Address = "Bangkok"
	i := Item{ProductID: "30000000-0000-4000-8000-000000000001", Quantity: 1}
	c.Items = []Item{i}
	if e := ValidateQuote(c); e != nil {
		t.Fatal(e)
	}
	c.Items = append(c.Items, i)
	if e := ValidateQuote(c); !errors.Is(e, shared.ErrInvalid) {
		t.Fatal(e)
	}
}
