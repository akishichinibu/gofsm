package gofsm_test

import (
	"errors"
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
	efsm, err := gofsm.NewEFSMWithContext(
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
		next, err := efsm.Transit(ctx, current, s.by)
		require.NoError(t, err)
		require.Equal(t, s.expected, next, "step %d: expected %v, got %v", i, s.expected, next)
		current = next
	}
}

func TestIllegalTransition(t *testing.T) {
	efsm, err := gofsm.NewEFSMWithContext(func(b gofsm.EFSMWithContextBuilder[OrderContext, OrderState, OrderEvent]) {
		b.From(StatePending).On(EventPay).ToConst(StatePaid)
	})
	require.NoError(t, err)

	ctx := OrderContext{}
	_, err = efsm.Transit(ctx, StatePending, EventShip)
	require.Error(t, err)

	var perr *gofsm.IllegalTransitError[OrderContext, OrderState, OrderEvent]
	if ok := errors.As(err, &perr); !ok {
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
	efsm, err := gofsm.NewEFSMWithContext(func(b gofsm.EFSMWithContextBuilder[OrderContext, OrderState, OrderEvent]) {
		b.From(StatePending).On(EventPay).ToConst(StatePaid)
	})
	require.NoError(t, err)

	ctx := OrderContext{}
	done := make(chan bool)

	for range 100 {
		go func() {
			defer func() { done <- true }()
			for range 1000 {
				_, _ = efsm.Transit(ctx, StatePending, EventPay)
			}
		}()
	}

	for range 100 {
		<-done
	}
}

type Context struct {
	Bais int
}

type Operation interface {
	MyOperation()
}

type Add struct {
	Diff int
}

var _ Operation = &Add{}

func (a *Add) MyOperation() {}

type Sub struct{}

var _ Operation = &Sub{}

func (s *Sub) MyOperation() {}

func TestEFSMWithParams(t *testing.T) {
	machine, err := gofsm.NewEFSMWithContext(func(builder gofsm.EFSMWithContextBuilder[Context, int, Operation]) {
		builder.From(1).On(&Add{Diff: 10}).To(func(ctx Context, from int, op Operation) (int, error) {
			add := op.(*Add)
			return from + ctx.Bais + add.Diff, nil
		})

		builder.From(20).On(&Sub{}).To(func(ctx Context, from int, op Operation) (int, error) {
			return from + ctx.Bais - 5, nil
		})
	})

	require.NoError(t, err)

	{
		state, err := machine.Transit(Context{Bais: 5}, 1, &Add{Diff: 10})
		require.NoError(t, err)
		require.Equal(t, 16, state)
	}

	{
		state, err := machine.Transit(Context{Bais: 3}, 20, &Sub{})
		require.NoError(t, err)
		require.Equal(t, 18, state)
	}

	_, err = machine.Transit(Context{Bais: 3}, 10, &Sub{})
	require.Error(t, err)
}
