package gofsm

import (
	"fmt"
	"sync"
)

type Transition[C any, S comparable, O comparable] func(context C, from S, by O) (to S, err error)

// MARK: EFSMWithContext
type EFSMWithContext[C any, S comparable, O comparable] interface {
	Transit(context C, from S, by O) (to S, err error)
}

type efsmWithContext[C any, S comparable, O comparable] struct {
	mutex sync.RWMutex
	table []rule[C, S, O]
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

func (e *efsmWithContext[C, S, O]) addRule(from S, by O, guard Guard[C, S, O], t Transition[C, S, O]) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, r := range e.table {
		if equalState(r.from, from) && equalOp(r.by, by) {
			panic(&illegalTransitDefinitionError[C, S, O]{
				Reason: fmt.Sprintf("Duplicated rule from %v by %v", from, by),
			})
		}
	}

	rule := newRule(from, by, guard, t)
	e.table = append(e.table, rule)
}

func newESFMWithContext[C any, S comparable, O comparable]() *efsmWithContext[C, S, O] {
	return &efsmWithContext[C, S, O]{
		mutex: sync.RWMutex{},
		table: make([]rule[C, S, O], 0),
	}
}

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
	b := &efsmWithContextBuilder[C, S, O]{
		m: newESFMWithContext[C, S, O](),
	}
	builderFunc(b)
	return b.m, nil
}

// MARK: Builder
type EFSMWithContextBuilder[C any, S comparable, O comparable] interface {
	From(from S) EFSMWithContextFromBuilder[C, S, O]
}

// MARK: FromBuilder
type EFSMWithContextFromBuilder[C any, S comparable, O comparable] interface {
	On(by O) EFSMWithContextOnBuilder[C, S, O]
	OnFunc(fromBuilder func(bfrom EFSMWithContextFromBuilder[C, S, O]))
}

// MARK: OnBuilder
type EFSMWithContextOnBuilder[C any, S comparable, O comparable] interface {
	To(f Transition[C, S, O])
	ToConst(v S)
	If(g Guard[C, S, O]) EFSMWithContextGuardBuilder[C, S, O]
}

// MARK: GuardBuilder
type EFSMWithContextGuardBuilder[C any, S comparable, O comparable] interface {
	To(f Transition[C, S, O])
	ToConst(v S)
}

// MARK: Impl

// MARK: EFSMWithContextBuilder
type efsmWithContextBuilder[C any, S comparable, O comparable] struct {
	m *efsmWithContext[C, S, O]
}

func (eb *efsmWithContextBuilder[C, S, B]) From(from S) EFSMWithContextFromBuilder[C, S, B] {
	return &efsmWithContextFromBuilder[C, S, B]{
		m:    eb.m,
		from: from,
	}
}

// MARK: EFSMWithContextFromBuilder
type efsmWithContextFromBuilder[C any, S comparable, O comparable] struct {
	m    *efsmWithContext[C, S, O]
	from S
}

func (f *efsmWithContextFromBuilder[C, S, O]) On(by O) EFSMWithContextOnBuilder[C, S, O] {
	return &efsmWithContextOnBuilder[C, S, O]{
		m:    f.m,
		from: f.from,
		on:   by,
	}
}

func (f *efsmWithContextFromBuilder[C, S, O]) OnFunc(bf func(b EFSMWithContextFromBuilder[C, S, O])) {
	bf(f)
}

// MARK: EFSMWithContextOnBuilder
type efsmWithContextOnBuilder[C any, S comparable, O comparable] struct {
	m    *efsmWithContext[C, S, O]
	from S
	on   O
}

func (eb *efsmWithContextOnBuilder[C, S, O]) If(g Guard[C, S, O]) EFSMWithContextGuardBuilder[C, S, O] {
	return &efsmWithContextGuardBuilder[C, S, O]{
		m:     eb.m,
		from:  eb.from,
		on:    eb.on,
		guard: g,
	}
}

func (eb *efsmWithContextOnBuilder[C, S, B]) To(f Transition[C, S, B]) {
	eb.If(nil).To(f)
}

func (eb *efsmWithContextOnBuilder[C, S, B]) ToConst(v S) {
	eb.If(nil).ToConst(v)
}

// MARK: EFSMWithContextGuardBuilder
type efsmWithContextGuardBuilder[C any, S comparable, O comparable] struct {
	m     *efsmWithContext[C, S, O]
	from  S
	on    O
	guard Guard[C, S, O]
}

func (eb *efsmWithContextGuardBuilder[C, S, O]) To(f Transition[C, S, O]) {
	eb.m.addRule(eb.from, eb.on, eb.guard, f)
}

func (eb *efsmWithContextGuardBuilder[C, S, O]) ToConst(v S) {
	eb.To(func(context C, from S, By O) (to S, err error) {
		return v, nil
	})
}
