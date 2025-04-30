package gofsm

import (
	"fmt"
	"reflect"
)

type rule[C any, S comparable, O comparable] struct {
	from  S
	by    O
	guard Guard[C, S, O]
	t     Transition[C, S, O]
}

func isPrimitiveKind(v reflect.Value) bool {
	return false ||
		v.Kind() == reflect.Bool ||
		v.Kind() == reflect.Int || v.Kind() == reflect.Int8 ||
		v.Kind() == reflect.Int16 || v.Kind() == reflect.Int32 ||
		false
}

func isValidState[S comparable](state S) bool {
	anyState := any(state)
	if _, ok := anyState.(fmt.Stringer); ok {
		return true
	}
	rv := reflect.ValueOf(state)
	return isPrimitiveKind(rv)
}

func newRule[C any, S comparable, O comparable](from S, by O, guard Guard[C, S, O], t Transition[C, S, O]) rule[C, S, O] {
	if !isValidState(from) {
		panic(&illegalTransitDefinitionError[C, S, O]{
			Reason: fmt.Sprintf("Illegal state %T(%v)", from, from),
		})
	}
	if t == nil {
		panic(&illegalTransitDefinitionError[C, S, O]{
			Reason: fmt.Sprintf("Illegal transition %T(%v)", t, t),
		})
	}
	if guard == nil {
		guard = trueGuard[C, S, O]
	}
	return rule[C, S, O]{
		from:  from,
		by:    by,
		guard: guard,
		t:     t,
	}
}

func extractState[S comparable](state S) any {
	anyState := any(state)
	if tv, ok := anyState.(fmt.Stringer); ok {
		return tv.String()
	}
	return state
}

func equalState[S comparable](a, b S) bool {
	vA := extractState(a)
	vB := extractState(b)
	return vA == vB
}

func indirectValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}

func equalOp[O comparable](op1, op2 O) bool {
	v1 := indirectValue(reflect.ValueOf(op1))
	v2 := indirectValue(reflect.ValueOf(op2))

	return v1.Type() == v2.Type() && reflect.DeepEqual(v1.Interface(), v2.Interface())
}
