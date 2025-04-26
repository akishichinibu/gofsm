package gofsm

import (
	"github.com/puzpuzpuz/xsync/v3"
)

type EFSM[C any, A comparable, B comparable] interface {
	Transit(context C, from A, by B) (to A, err error)
}

type efsm[C any, A comparable, B comparable] struct {
	table *xsync.MapOf[A, *xsync.MapOf[B, Transition[C, A, B]]]
}

func newEFSM[C any, A comparable, B comparable]() *efsm[C, A, B] {
	return &efsm[C, A, B]{
		table: xsync.NewMapOf[A, *xsync.MapOf[B, Transition[C, A, B]]](),
	}
}

func (e *efsm[C, A, B]) Transit(context C, from A, by B) (to A, err error) {
	fromTable, ok := e.table.Load(from)
	if !ok {
		return from, &IllegalTransitError[C, A, B]{m: e, From: from, By: by}
	}
	tf, ok := fromTable.Load(by)
	if !ok {
		return from, &IllegalTransitError[C, A, B]{m: e, From: from, By: by}
	}
	return tf(context, from, by)
}
