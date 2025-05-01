package gofsm

import (
	"fmt"
	"sync"
)

// Transition defines a function type for transitioning between states.
// It takes a context, a current state, and an operation, and returns the next state and an error.
type Transition[C any, S comparable, O comparable] func(context C, from S, operation O) (to S, err error)

// EFSMWithContext is an interface for an extended finite state machine with context.
// It provides a method to perform state transitions.
type EFSMWithContext[C any, S comparable, O comparable] interface {
	Transit(context C, from S, operation O) (to S, err error)
}

// efsmWithContext is the implementation of EFSMWithContext.
// It uses a mutex for thread safety and maintains a table of transition rules.
type efsmWithContext[C any, S comparable, O comparable] struct {
	mutex sync.RWMutex
	table []rule[C, S, O]
}

// Transit performs a state transition based on the current state, operation, and context.
// It checks the transition rules and applies the first matching rule with a valid guard.
func (e *efsmWithContext[C, S, O]) Transit(context C, from S, operation O) (to S, err error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	matched := false

	for _, rule := range e.table {
		if equalState(rule.from, from) {
			if equalOp(rule.operation, operation) {
				matched = true

				if rule.guard(context, from, operation) {
					return rule.t(context, from, operation)
				}
			}
		}
	}

	// If no valid transition is found, return an error.
	er := &IllegalTransitError[C, S, O]{
		m:         e,
		From:      from,
		Operation: operation,
	}

	if matched {
		er.Reason = "guard failed"
	} else {
		er.Reason = "no matching rule"
	}

	return from, er
}

// addRule adds a new transition rule to the state machine.
// It ensures that no duplicate rules exist for the same state and operation.
func (e *efsmWithContext[C, S, O]) addRule(from S, operation O, guard Guard[C, S, O], t Transition[C, S, O]) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, r := range e.table {
		if equalState(r.from, from) && equalOp(r.operation, operation) {
			panic(&illegalTransitDefinitionError[C, S, O]{
				Reason: fmt.Sprintf("Duplicated rule from %v operation %v", from, operation),
			})
		}
	}

	rule := newRule(from, operation, guard, t)
	e.table = append(e.table, rule)
}

// newESFMWithContext creates a new instance of efsmWithContext.
func newESFMWithContext[C any, S comparable, O comparable]() *efsmWithContext[C, S, O] {
	return &efsmWithContext[C, S, O]{
		mutex: sync.RWMutex{},
		table: make([]rule[C, S, O], 0),
	}
}

// NewEFSMWithContext creates a new EFSMWithContext using a builder function.
// It recovers from any illegal transition definition errors during the building process.
func NewEFSMWithContext[C any, S comparable, O comparable](builderFunc func(b EFSMWithContextBuilder[C, S, O])) (m EFSMWithContext[C, S, O], err error) {
	defer func() {
		if expt := recover(); expt != nil {
			switch te := expt.(type) {
			case *illegalTransitDefinitionError[C, S, O]:
				m = nil
				err = te
			default:
				panic(expt)
			}
		}
	}()

	builder := &efsmWithContextBuilder[C, S, O]{
		m: newESFMWithContext[C, S, O](),
	}
	builderFunc(builder)

	return builder.m, nil
}

// EFSMWithContextBuilder is an interface for building an EFSMWithContext.
// It allows specifying the starting state for a transition.
type EFSMWithContextBuilder[C any, S comparable, O comparable] interface {
	From(from S) EFSMWithContextFromBuilder[C, S, O]
}

// EFSMWithContextFromBuilder is an interface for specifying the operation for a transition.
type EFSMWithContextFromBuilder[C any, S comparable, O comparable] interface {
	On(operation O) EFSMWithContextOnBuilder[C, S, O]
	OnFunc(fromBuilder func(bfrom EFSMWithContextFromBuilder[C, S, O]))
}

// EFSMWithContextOnBuilder is an interface for specifying the target state or transition function.
type EFSMWithContextOnBuilder[C any, S comparable, O comparable] interface {
	To(f Transition[C, S, O])
	ToConst(v S)
	If(g Guard[C, S, O]) EFSMWithContextGuardBuilder[C, S, O]
}

// EFSMWithContextGuardBuilder is an interface for specifying the target state or transition function with a guard condition.
type EFSMWithContextGuardBuilder[C any, S comparable, O comparable] interface {
	To(f Transition[C, S, O])
	ToConst(v S)
}

// efsmWithContextBuilder is the implementation of EFSMWithContextBuilder.
type efsmWithContextBuilder[C any, S comparable, O comparable] struct {
	m *efsmWithContext[C, S, O]
}

// From specifies the starting state for a transition.
func (eb *efsmWithContextBuilder[C, S, B]) From(from S) EFSMWithContextFromBuilder[C, S, B] {
	return &efsmWithContextFromBuilder[C, S, B]{
		m:    eb.m,
		from: from,
	}
}

// efsmWithContextFromBuilder is the implementation of EFSMWithContextFromBuilder.
type efsmWithContextFromBuilder[C any, S comparable, O comparable] struct {
	m    *efsmWithContext[C, S, O]
	from S
}

// On specifies the operation for a transition.
func (f *efsmWithContextFromBuilder[C, S, O]) On(operation O) EFSMWithContextOnBuilder[C, S, O] {
	return &efsmWithContextOnBuilder[C, S, O]{
		m:    f.m,
		from: f.from,
		on:   operation,
	}
}

// OnFunc allows specifying multiple operations for a transition using a function.
func (f *efsmWithContextFromBuilder[C, S, O]) OnFunc(bf func(b EFSMWithContextFromBuilder[C, S, O])) {
	bf(f)
}

// efsmWithContextOnBuilder is the implementation of EFSMWithContextOnBuilder.
type efsmWithContextOnBuilder[C any, S comparable, O comparable] struct {
	m    *efsmWithContext[C, S, O]
	from S
	on   O
}

// If specifies a guard condition for the transition.
func (eb *efsmWithContextOnBuilder[C, S, O]) If(g Guard[C, S, O]) EFSMWithContextGuardBuilder[C, S, O] {
	return &efsmWithContextGuardBuilder[C, S, O]{
		m:     eb.m,
		from:  eb.from,
		on:    eb.on,
		guard: g,
	}
}

// To specifies a transition function for the transition.
func (eb *efsmWithContextOnBuilder[C, S, B]) To(f Transition[C, S, B]) {
	eb.If(nil).To(f)
}

// ToConst specifies a constant target state for the transition.
func (eb *efsmWithContextOnBuilder[C, S, B]) ToConst(v S) {
	eb.If(nil).ToConst(v)
}

// efsmWithContextGuardBuilder is the implementation of EFSMWithContextGuardBuilder.
type efsmWithContextGuardBuilder[C any, S comparable, O comparable] struct {
	m     *efsmWithContext[C, S, O]
	from  S
	on    O
	guard Guard[C, S, O]
}

// To specifies a transition function for the transition with a guard condition.
func (eb *efsmWithContextGuardBuilder[C, S, O]) To(f Transition[C, S, O]) {
	eb.m.addRule(eb.from, eb.on, eb.guard, f)
}

// ToConst specifies a constant target state for the transition with a guard condition.
func (eb *efsmWithContextGuardBuilder[C, S, O]) ToConst(v S) {
	eb.To(func(_ C, _ S, _ O) (S, error) {
		return v, nil
	})
}
