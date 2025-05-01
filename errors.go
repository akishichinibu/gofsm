package gofsm

import "fmt"

type illegalTransitDefinitionError[C any, S comparable, O comparable] struct {
	Reason string
}

func (e *illegalTransitDefinitionError[C, A, B]) Error() string {
	return fmt.Sprintf("Illegal transit definition: %s", e.Reason)
}

type IllegalTransitError[C any, S comparable, O comparable] struct {
	m         EFSMWithContext[C, S, O]
	From      S
	Operation O
	Reason    string
}

func (e *IllegalTransitError[C, A, B]) Error() string {
	return fmt.Sprintf("illeagal transit in %T, from %T(%v) by %T(%v): %s", e.m, e.From, e.From, e.Operation, e.Operation, e.Reason)
}
