package gofsm_test

import (
	"testing"

	"github.com/akishichinibu/gofsm"
	"github.com/stretchr/testify/require"
)

func TestSimpleFSM(t *testing.T) {
	type State string
	type Event string

	const (
		Idle    State = "idle"
		Running State = "running"
		Done    State = "done"

		Start  Event = "start"
		Finish Event = "finish"
	)

	fsm, err := gofsm.NewFSM(func(b gofsm.FSMBuilder[State, Event]) {
		b.From(Idle).On(Start).To(Running)
		b.From(Running).On(Finish).To(Done)
	})
	require.NoError(t, err)

	state, err := fsm.Transit(Idle, Start)
	require.NoError(t, err)
	require.Equal(t, Running, state)

	state, err = fsm.Transit(state, Finish)
	require.NoError(t, err)
	require.Equal(t, Done, state)

	_, err = fsm.Transit(Idle, Finish)
	require.Error(t, err)
}
