package gofsm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type testStruct struct{}

var _ fmt.Stringer = testStruct{}

func (t testStruct) String() string {
	return "abc"
}

func TestExtraState(t *testing.T) {
	t.Run("simple type", func(t *testing.T) {
		v := extractState(1)
		require.Equal(t, "int", fmt.Sprintf("%T", v))
	})

	t.Run("string type", func(t *testing.T) {
		v := extractState(testStruct{})
		require.Equal(t, "abc", v)
	})

}
