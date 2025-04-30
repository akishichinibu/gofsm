package gofsm_test

import (
	"testing"

	"github.com/akishichinibu/gofsm"
	"github.com/stretchr/testify/require"
)

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
	machine, err := gofsm.NewEFSMWithContext(func(b gofsm.EFSMWithContextBuilder[Context, int, Operation]) {

		b.From(1).On(&Add{Diff: 10}).To(func(ctx Context, from int, op Operation) (int, error) {
			add := op.(*Add)
			return from + ctx.Bais + add.Diff, nil
		})

		b.From(20).On(&Sub{}).To(func(ctx Context, from int, op Operation) (int, error) {
			return from + ctx.Bais - 5, nil
		})

	})

	require.NoError(t, err)

	to, err := machine.Transit(Context{Bais: 5}, 1, &Add{Diff: 10})
	require.NoError(t, err)
	require.Equal(t, 16, to)

	to, err = machine.Transit(Context{Bais: 3}, 20, &Sub{})
	require.NoError(t, err)
	require.Equal(t, 18, to)

	_, err = machine.Transit(Context{Bais: 3}, 10, &Sub{})
	require.Error(t, err)
}
