package file

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilelock(t *testing.T) {
	locked := TryLock("./go-lock.lock")
	require.True(t, locked)
	locked2 := New().TryLock("./go-lock.lock")
	require.False(t, locked2)
	Unlock("./go-lock.lock")

}
