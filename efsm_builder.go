package gofsm

import (
	"fmt"
)

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
		m: newEFSM[C, S, O](),
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
	eb.m.mutex.Lock()
	defer eb.m.mutex.Unlock()

	for _, r := range eb.m.table {
		if equalState(r.from, eb.from) && equalOp(r.by, eb.on) {
			panic(&illegalTransitDefinitionError[C, S, O]{
				Reason: fmt.Sprintf("Duplicated rule from %v by %v", eb.from, eb.on),
			})
		}
	}

	rule := newRule(eb.from, eb.on, eb.guard, f)
	eb.m.table = append(eb.m.table, rule)
}

func (eb *efsmWithContextGuardBuilder[C, S, O]) ToConst(v S) {
	eb.To(func(context C, from S, By O) (to S, err error) {
		return v, nil
	})
}
