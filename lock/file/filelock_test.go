package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilelock(t *testing.T) {
	os.MkdirAll("./lock", 0755)
	locked := TryLock("./lock/go-lock.lock")
	require.True(t, locked)
	locked2 := New().TryLock("./lock/go-lock.lock")
	require.False(t, locked2)
	Unlock("./go-lock.lock")

}
