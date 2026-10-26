package state_machine

import (
	"context"
	"testing"

	tu "github.com/siper92/akha/lang/tests_utils"
)

type orderState string

const (
	stReceived  orderState = "received"
	stValidate  orderState = "validate"
	stReserve   orderState = "reserve"
	stBackorder orderState = "backorder"
	stPay       orderState = "pay"
	stRetry     orderState = "retry"
	stShip      orderState = "ship"
	stDelivered orderState = "delivered"
	stCancelled orderState = "cancelled"
	stRejected  orderState = "rejected"
)

type order struct {
	Items, Stock, Restock          int
	PayFails, Attempts, MaxRetries int
}

type orderInput struct {
	from orderState
	data order
}

type orderResult struct {
	State orderState
	Path  []string
}

func receiveOrder(_ context.Context, _ *order) (orderState, error) {
	return stValidate, nil
}

func validateOrder(_ context.Context, o *order) (orderState, error) {
	if o.Items <= 0 {
		return stRejected, nil
	}
	return stReserve, nil
}

func reserveStock(_ context.Context, o *order) (orderState, error) {
	if o.Stock < o.Items {
		return stBackorder, nil
	}
	o.Stock -= o.Items
	return stPay, nil
}

func waitRestock(_ context.Context, o *order) (orderState, error) {
	if o.Restock == 0 {
		return stCancelled, nil
	}
	o.Stock += o.Restock
	o.Restock = 0
	return stReserve, nil
}

func chargePayment(_ context.Context, o *order) (orderState, error) {
	o.Attempts++
	if o.Attempts <= o.PayFails {
		return stRetry, nil
	}
	return stShip, nil
}

func retryPayment(_ context.Context, o *order) (orderState, error) {
	if o.Attempts > o.MaxRetries {
		return stCancelled, nil
	}
	return stPay, nil
}

func shipOrder(_ context.Context, _ *order) (orderState, error) {
	return stDelivered, nil
}

// transitions:
//   - received  -> validate             always
//   - validate  -> reserve | rejected   rejected when the order has no items
//   - reserve   -> pay | backorder      backorder when stock is lower than items
//   - backorder -> reserve | cancelled  reserve again after a restock, cancel without one
//   - pay       -> ship | retry         retry while attempts are within PayFails
//   - retry     -> pay | cancelled      cancel once attempts exceed MaxRetries
//   - ship      -> delivered            always
func orderMachine() *Machine[order, orderState] {
	return NewMachine[order, orderState](stReceived).
		Step(stReceived, receiveOrder, stValidate).
		Step(stValidate, validateOrder, stReserve, stRejected).
		Step(stReserve, reserveStock, stPay, stBackorder).
		Step(stBackorder, waitRestock, stReserve, stCancelled).
		Step(stPay, chargePayment, stShip, stRetry).
		Step(stRetry, retryPayment, stPay, stCancelled).
		Step(stShip, shipOrder, stDelivered).
		Final(stDelivered, stCancelled, stRejected)
}

// --- order flow

func TestOrderMachine_Flow(t *testing.T) {
	cases := []tu.Case[orderInput, orderResult]{
		{
			Name:  "happy_path",
			Input: orderInput{data: order{Items: 2, Stock: 5}},
			Expected: orderResult{State: stDelivered, Path: []string{
				"received>validate", "validate>reserve", "reserve>pay", "pay>ship", "ship>delivered",
			}},
		},
		{
			Name:  "empty_order_rejected",
			Input: orderInput{data: order{Items: 0, Stock: 5}},
			Expected: orderResult{State: stRejected, Path: []string{
				"received>validate", "validate>rejected",
			}},
		},
		{
			Name:  "backorder_restocked",
			Input: orderInput{data: order{Items: 3, Stock: 1, Restock: 5}},
			Expected: orderResult{State: stDelivered, Path: []string{
				"received>validate", "validate>reserve", "reserve>backorder", "backorder>reserve",
				"reserve>pay", "pay>ship", "ship>delivered",
			}},
		},
		{
			Name:  "backorder_cancelled",
			Input: orderInput{data: order{Items: 3, Stock: 1}},
			Expected: orderResult{State: stCancelled, Path: []string{
				"received>validate", "validate>reserve", "reserve>backorder", "backorder>cancelled",
			}},
		},
		{
			Name:  "payment_retry_succeeds",
			Input: orderInput{data: order{Items: 1, Stock: 1, PayFails: 1, MaxRetries: 2}},
			Expected: orderResult{State: stDelivered, Path: []string{
				"received>validate", "validate>reserve", "reserve>pay",
				"pay>retry", "retry>pay", "pay>ship", "ship>delivered",
			}},
		},
		{
			Name:  "payment_retries_exhausted",
			Input: orderInput{data: order{Items: 1, Stock: 1, PayFails: 3, MaxRetries: 1}},
			Expected: orderResult{State: stCancelled, Path: []string{
				"received>validate", "validate>reserve", "reserve>pay",
				"pay>retry", "retry>pay", "pay>retry", "retry>cancelled",
			}},
		},
		{
			Name:  "resume_from_pay",
			Input: orderInput{from: stPay, data: order{Items: 1}},
			Expected: orderResult{State: stDelivered, Path: []string{
				"pay>ship", "ship>delivered",
			}},
		},
	}

	tu.Run(tu.New(t), cases, func(in orderInput) (orderResult, error) {
		var path []string
		m := orderMachine()
		m.OnStep = func(tr Transition[orderState]) {
			path = append(path, string(tr.From)+">"+string(tr.To))
		}

		from := in.from
		if from == "" {
			from = stReceived
		}
		state, err := m.RunFrom(context.Background(), from, &in.data)
		return orderResult{State: state, Path: path}, err
	}, nil)
}
