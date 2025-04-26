package gofsm

import "fmt"

type illegalTransitDefinitionError[C any, A comparable, B comparable] struct {
	Reason string
}

func (e *illegalTransitDefinitionError[C, A, B]) Error() string {
	return fmt.Sprintf("Illegal transit definition: %s", e.Reason)
}

type IllegalTransitError[C any, A comparable, B comparable] struct {
	m    EFSM[C, A, B]
	From A
	By   B
}

func (e *IllegalTransitError[C, A, B]) Error() string {
	return fmt.Sprintf("illeagal transit in %T, from %T(%v) by %T(%v)", e.m, e.From, e.From, e.By, e.By)
}
