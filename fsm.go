package gofsm

type FSM[S comparable, O comparable] interface {
	Transit(from S, by O) (to S, err error)
}

type fsm[S comparable, O comparable] struct {
	m *efsmWithContext[any, S, O]
}

func (f *fsm[S, O]) Transit(from S, by O) (to S, err error) {
	return f.m.Transit(nil, from, by)
}

func NewFSM[S comparable, O comparable](builderFunc func(b FSMBuilder[S, O])) (FSM[S, O], error) {
	b := &fsmBuilder[S, O]{
		m: &fsm[S, O]{
			m: newESFMWithContext[any, S, O](),
		},
	}
	builderFunc(b)
	return b.m, nil
}

// MARK: Builder
type FSMBuilder[S comparable, O comparable] interface {
	From(from S) FSMFromBuilder[S, O]
}

type FSMFromBuilder[S comparable, O comparable] interface {
	On(by O) FSMOnBuilder[S, O]
	OnFunc(fromBuilder func(bfrom FSMFromBuilder[S, O]))
}

type FSMOnBuilder[S comparable, O comparable] interface {
	To(v S)
}

// MARK: Impl

type fsmBuilder[S comparable, O comparable] struct {
	m *fsm[S, O]
}

func (fb *fsmBuilder[S, O]) From(from S) FSMFromBuilder[S, O] {
	return &fsmFromBuilder[S, O]{
		m:    fb.m,
		from: from,
	}
}

type fsmFromBuilder[S comparable, O comparable] struct {
	m    *fsm[S, O]
	from S
}

func (f *fsmFromBuilder[S, O]) On(by O) FSMOnBuilder[S, O] {
	return &fsmOnBuilder[S, O]{
		m:    f.m,
		from: f.from,
		on:   by,
	}
}

func (f *fsmFromBuilder[S, O]) OnFunc(bf func(b FSMFromBuilder[S, O])) {
	bf(f)
}

type fsmOnBuilder[S comparable, O comparable] struct {
	m    *fsm[S, O]
	from S
	on   O
}

func (f *fsmOnBuilder[S, O]) To(to S) {
	f.m.m.addRule(f.from, f.on, nil, func(context any, from S, by O) (S, error) {
		return to, nil
	})
}
