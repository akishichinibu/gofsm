package gofsm

import (
	"fmt"

	"github.com/puzpuzpuz/xsync/v3"
)

type Transition[C any, A comparable, B comparable] func(context C, from A, by B) (to A, err error)

func NewEFSM[C any, A comparable, B comparable](builderFunc func(b EFSMBuilder[C, A, B])) (m EFSM[C, A, B], err error) {
	defer func() {
		if expt := recover(); expt != nil {
			switch te := expt.(type) {
			case *illegalTransitDefinitionError[C, A, B]:
				m = nil
				err = te
			default:
				panic(expt)
			}
		}
	}()
	b := &efsmBuilder[C, A, B]{
		m: newEFSM[C, A, B](),
	}
	builderFunc(b)
	return b.m, nil
}

// MARK: EFSMBuilder
type EFSMBuilder[C any, A comparable, B comparable] interface {
	From(from A) EFSMFromBuilder[C, A, B]
}

type efsmBuilder[C any, A comparable, B comparable] struct {
	m *efsm[C, A, B]
}

func (eb *efsmBuilder[C, A, B]) From(from A) EFSMFromBuilder[C, A, B] {
	return &efsmFromBuilder[C, A, B]{
		m:    eb.m,
		from: from,
	}
}

// MARK: EFSMFromBuilder
type EFSMFromBuilder[C any, A comparable, B comparable] interface {
	On(by B) EFSMOnBuilder[C, A, B]
	OnFunc(fromBuilder func(bfrom EFSMFromBuilder[C, A, B]))
}

type efsmFromBuilder[C any, A comparable, B comparable] struct {
	m    *efsm[C, A, B]
	from A
}

func (f *efsmFromBuilder[C, A, B]) On(by B) EFSMOnBuilder[C, A, B] {
	return &efsmOnBuilder[C, A, B]{
		m:    f.m,
		from: f.from,
		on:   by,
	}
}

func (f *efsmFromBuilder[C, A, B]) OnFunc(bf func(b EFSMFromBuilder[C, A, B])) {
	bf(f)
}

// MARK: EFSMOnBuilder
type EFSMOnBuilder[C any, A comparable, B comparable] interface {
	To(f Transition[C, A, B])
	ToConst(v A)
}

type efsmOnBuilder[C any, A comparable, B comparable] struct {
	m    *efsm[C, A, B]
	from A
	on   B
}

func (eb *efsmOnBuilder[C, A, B]) To(f Transition[C, A, B]) {
	toTable, loaded := eb.m.table.LoadOrCompute(
		eb.from,
		func() *xsync.MapOf[B, Transition[C, A, B]] {
			return xsync.NewMapOf[B, Transition[C, A, B]]()
		},
	)
	if !loaded {
		eb.m.table.Store(eb.from, toTable)
	}
	_, ok := toTable.LoadOrStore(eb.on, f)
	if ok {
		panic(&illegalTransitDefinitionError[C, A, B]{
			Reason: fmt.Sprintf("Duplicated rule from %v by %v", eb.from, eb.on),
		})
	}
}

func (eb *efsmOnBuilder[C, A, B]) ToConst(v A) {
	eb.To(func(context C, from A, by B) (to A, err error) {
		return v, nil
	})
}
