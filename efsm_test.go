package gofsm_test

import (
	"testing"

	"github.com/akishichinibu/gofsm"
	"github.com/stretchr/testify/require"
)

type OrderState string
type OrderEvent string

const (
	StatePending   OrderState = "Pending"
	StatePaid      OrderState = "Paid"
	StateShipped   OrderState = "Shipped"
	StateDelivered OrderState = "Delivered"
	StateCancelled OrderState = "Cancelled"

	EventPay     OrderEvent = "Pay"
	EventShip    OrderEvent = "Ship"
	EventDeliver OrderEvent = "Deliver"
	EventCancel  OrderEvent = "Cancel"
)

type OrderContext struct{}

func TestOrderStateMachine(t *testing.T) {
	sm, err := gofsm.NewEFSMWithContext(
		func(b gofsm.EFSMWithContextBuilder[OrderContext, OrderState, OrderEvent]) {
			b.From(StatePending).On(EventPay).ToConst(StatePaid)
			b.From(StatePending).On(EventCancel).ToConst(StateCancelled)
			b.From(StatePaid).On(EventShip).ToConst(StateShipped)
			b.From(StatePaid).On(EventCancel).ToConst(StateCancelled)
			b.From(StateShipped).On(EventDeliver).ToConst(StateDelivered)
		},
	)
	require.NoError(t, err)

	ctx := OrderContext{}

	type step struct {
		from     OrderState
		by       OrderEvent
		expected OrderState
	}
	steps := []step{
		{StatePending, EventPay, StatePaid},
		{StatePaid, EventShip, StateShipped},
		{StateShipped, EventDeliver, StateDelivered},
	}

	current := StatePending
	for i, s := range steps {
		next, err := sm.Transit(ctx, current, s.by)
		require.NoError(t, err)
		require.Equal(t, s.expected, next, "step %d: expected %v, got %v", i, s.expected, next)
		current = next
	}
}

func TestIllegalTransition(t *testing.T) {
	sm, err := gofsm.NewEFSMWithContext(func(b gofsm.EFSMWithContextBuilder[OrderContext, OrderState, OrderEvent]) {
		b.From(StatePending).On(EventPay).ToConst(StatePaid)
	})
	require.NoError(t, err)

	ctx := OrderContext{}
	_, err = sm.Transit(ctx, StatePending, EventShip)
	require.Error(t, err)

	if _, ok := err.(*gofsm.IllegalTransitError[OrderContext, OrderState, OrderEvent]); !ok {
		require.Fail(t, "expected IllegalTransitError, got %T", err)
	}
}

func TestDuplicateDefinitionPanic(t *testing.T) {
	_, err := gofsm.NewEFSMWithContext(func(b gofsm.EFSMWithContextBuilder[OrderContext, OrderState, OrderEvent]) {
		b.From(StatePending).On(EventPay).ToConst(StatePaid)
		b.From(StatePending).On(EventPay).ToConst(StatePaid)
	})
	require.Error(t, err)
}

func TestConcurrentTransit(t *testing.T) {
	sm, err := gofsm.NewEFSMWithContext(func(b gofsm.EFSMWithContextBuilder[OrderContext, OrderState, OrderEvent]) {
		b.From(StatePending).On(EventPay).ToConst(StatePaid)
	})
	require.NoError(t, err)

	ctx := OrderContext{}
	done := make(chan bool)

	for i := 0; i < 100; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 1000; j++ {
				_, _ = sm.Transit(ctx, StatePending, EventPay)
			}
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}
