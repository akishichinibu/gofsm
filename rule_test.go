//nolint:testpackage
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

type testStruct2 struct{}

type testString string

func TestIsValidState(t *testing.T) {
	t.Run("primitive type", func(t *testing.T) {
		require.True(t, isValidState(1))
	})

	t.Run("defined type", func(t *testing.T) {
		require.True(t, isValidState(testString("abc")))
	})

	t.Run("string type", func(t *testing.T) {
		require.True(t, isValidState(testStruct{}))
	})

	t.Run("invalid type", func(t *testing.T) {
		require.False(t, isValidState(testStruct2{}))
	})
}

func TestExtraState(t *testing.T) {
	t.Run("simple type", func(t *testing.T) {
		v := extractState(1)
		require.Equal(t, "int", fmt.Sprintf("%T", v))
	})

	t.Run("defined type", func(t *testing.T) {
		v := extractState(testString("abc"))
		require.NotEqual(t, "abc", v)
		require.Equal(t, "gofsm.testString", fmt.Sprintf("%T", v))
		require.Equal(t, "abc", string(v.(testString)))
	})

	t.Run("string type", func(t *testing.T) {
		v := extractState(testStruct{})
		require.Equal(t, "abc", v)
	})
}
