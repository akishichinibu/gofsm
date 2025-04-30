package gofsm

type Guard[C any, S comparable, O comparable] func(context C, from S, by O) bool

func trueGuard[C any, S comparable, O comparable](context C, from S, by O) bool {
	return true
}
