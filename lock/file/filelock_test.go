package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilelock(t *testing.T) {
	os.MkdirAll("./lock", 0755)
	locked, err := TryLock("./lock/go-lock.lock")
	require.NoError(t, err)
	require.True(t, locked)
	locked2, err := New().TryLock("./lock/go-lock.lock")
	require.NoError(t, err)
	require.False(t, locked2)
	Unlock("./go-lock.lock")

}
