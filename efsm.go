package gofsm

import (
	"sync"
)

type Transition[C any, S comparable, O comparable] func(context C, from S, by O) (to S, err error)

type Guard[C any, S comparable, O comparable] func(context C, from S, by O) bool

func trueGuard[C any, S comparable, O comparable](context C, from S, by O) bool {
	return true
}

type EFSMWithContext[C any, S comparable, O comparable] interface {
	Transit(context C, from S, by O) (to S, err error)
}

type efsmWithContext[C any, S comparable, O comparable] struct {
	mutex sync.RWMutex
	table []rule[C, S, O]
}

func newEFSM[C any, S comparable, O comparable]() *efsmWithContext[C, S, O] {
	return &efsmWithContext[C, S, O]{
		mutex: sync.RWMutex{},
		table: make([]rule[C, S, O], 0),
	}
}

func (e *efsmWithContext[C, S, O]) Transit(context C, from S, by O) (to S, err error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	matched := false

	for _, r := range e.table {
		if equalState(r.from, from) {
			if equalOp(r.by, by) {
				matched = true
				if r.guard(context, from, by) {
					return r.t(context, from, by)
				}
			}
		}
	}

	er := &IllegalTransitError[C, S, O]{
		m:    e,
		From: from,
		By:   by,
	}

	if matched {
		er.Reason = "guard failed"
	} else {
		er.Reason = "no matching rule"
	}

	return from, er
}
