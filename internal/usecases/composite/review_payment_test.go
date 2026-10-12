package composite

import (
	"bakery/internal/domains/orders"
	portin "bakery/internal/ports/port_in/orders"
	catalogout "bakery/internal/ports/port_out/catalog"
	inventoryout "bakery/internal/ports/port_out/inventory"
	notificationout "bakery/internal/ports/port_out/notifications"
	ordersout "bakery/internal/ports/port_out/orders"
	paymentsout "bakery/internal/ports/port_out/payments"
	"bakery/internal/shared"
	"context"
	"errors"
	"testing"
)

const testID = "10000000-0000-4000-8000-000000000001"

type testScope struct{}

func (*testScope) TransactionScope() {}

type txRecorder struct {
	scope                 shared.Transaction
	committed, rolledBack bool
	commitErr             error
}

func (r *txRecorder) Exec(_ context.Context, f func(shared.Transaction) error) error {
	e := f(r.scope)
	if e != nil {
		r.rolledBack = true
		return e
	}
	if r.commitErr != nil {
		return r.commitErr
	}
	r.committed = true
	return nil
}

type recorder struct {
	t            *testing.T
	tx           shared.Transaction
	calls        []string
	fail         string
	err          error
	paymentState string
}

func (r *recorder) call(tx shared.Transaction, name string) error {
	r.t.Helper()
	if tx != r.tx {
		r.t.Fatal("repositories did not receive same transaction")
	}
	r.calls = append(r.calls, name)
	if r.fail == name {
		return r.err
	}
	return nil
}

type orderMock struct {
	ordersout.Repository
	r *recorder
}

func (m orderMock) LockStatus(_ context.Context, tx shared.Transaction, _ string) (string, error) {
	return "pending_verification", m.r.call(tx, "lock-order")
}
func (m orderMock) ItemCount(_ context.Context, tx shared.Transaction, _ string) (int, error) {
	return 1, m.r.call(tx, "item-count")
}
func (m orderMock) ChangeStatus(_ context.Context, tx shared.Transaction, _, _, _ string) error {
	return m.r.call(tx, "status")
}
func (m orderMock) History(_ context.Context, tx shared.Transaction, _, _, _, _, _ string) error {
	return m.r.call(tx, "history")
}
func (m orderMock) GetTx(_ context.Context, tx shared.Transaction, _ string) (orders.Data, error) {
	return orders.Data{ID: testID, Number: "TB-Test", Total: 100}, m.r.call(tx, "read")
}

type paymentMock struct {
	paymentsout.Repository
	r *recorder
}

func (m paymentMock) OrderID(context.Context, string) (string, error) { return testID, nil }
func (m paymentMock) LockStatus(_ context.Context, tx shared.Transaction, _ string) (string, error) {
	return m.r.paymentState, m.r.call(tx, "lock-payment")
}
func (m paymentMock) Approve(_ context.Context, tx shared.Transaction, _, _ string) error {
	return m.r.call(tx, "approve")
}

type inventoryMock struct {
	inventoryout.Repository
	r *recorder
}

func (m inventoryMock) Active(_ context.Context, tx shared.Transaction, _ string) ([]orders.Item, error) {
	return []orders.Item{{ProductID: testID, Quantity: 1}}, m.r.call(tx, "active")
}
func (m inventoryMock) RecordSale(_ context.Context, tx shared.Transaction, _ string, _ orders.Item, _ string) error {
	return m.r.call(tx, "movement")
}
func (m inventoryMock) Commit(_ context.Context, tx shared.Transaction, _ string) error {
	return m.r.call(tx, "inventory-commit")
}

type productMock struct {
	catalogout.Repository
	r *recorder
}

func (m productMock) Deduct(_ context.Context, tx shared.Transaction, _ string, _ int) error {
	return m.r.call(tx, "deduct")
}

type outboxMock struct {
	notificationout.Outbox
	r *recorder
}

func (m outboxMock) Enqueue(_ context.Context, tx shared.Transaction, _, _ string) error {
	return m.r.call(tx, "outbox")
}
func TestApprovalUsesOneTransactionAndPropagatesFailures(t *testing.T) {
	for _, tc := range []struct {
		name, fail, state string
		commitFailure     bool
	}{{"success", "", "submitted", false}, {"stock-failure", "deduct", "submitted", false}, {"outbox-failure", "outbox", "submitted", false}, {"commit-failure", "", "submitted", true}, {"idempotent", "", "approved", false}} {
		t.Run(tc.name, func(t *testing.T) {
			boom := errors.New("forced failure")
			scope := &testScope{}
			r := &recorder{t: t, tx: scope, fail: tc.fail, err: boom, paymentState: tc.state}
			tx := &txRecorder{scope: scope}
			if tc.commitFailure {
				tx.commitErr = boom
			}
			uc := NewApprovePaymentUsecase(ApprovePaymentArgs{ExecWithTx: tx, OrderRepository: orderMock{r: r}, PaymentRepository: paymentMock{r: r}, InventoryRepository: inventoryMock{r: r}, CatalogRepository: productMock{r: r}, Outbox: outboxMock{r: r}})
			_, e := uc.Handle(context.Background(), portin.ApprovePaymentCommand{PaymentID: testID, Actor: testID})
			if tc.fail != "" || tc.commitFailure {
				if !errors.Is(e, boom) || tx.committed {
					t.Fatal("failure not propagated", e)
				}
				if tc.fail != "" && !tx.rolledBack {
					t.Fatal("callback failure did not rollback")
				}
			} else if e != nil || !tx.committed {
				t.Fatal(e)
			}
			if tc.state == "approved" {
				for _, name := range r.calls {
					if name == "deduct" || name == "outbox" {
						t.Fatal("duplicate approval wrote", name)
					}
				}
			}
		})
	}
}
func TestInvalidQuoteDoesNotReachDependencies(t *testing.T) {
	uc := NewQuoteCheckoutUsecase(QuoteCheckoutArgs{})
	_, e := uc.Handle(context.Background(), portin.QuoteCheckoutCommand{})
	if !errors.Is(e, shared.ErrInvalid) {
		t.Fatal(e)
	}
}
