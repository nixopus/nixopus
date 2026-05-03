package ssh

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	resetInvalidateHooksForTest()
	InvalidateAllSSHManagerCaches()
	orig := sshIdleCleanupTickerInterval
	sshIdleCleanupTickerInterval = 25 * time.Millisecond
	code := m.Run()
	sshIdleCleanupTickerInterval = orig
	InvalidateAllSSHManagerCaches()
	resetInvalidateHooksForTest()
	os.Exit(code)
}

func TestAppendUnique(t *testing.T) {
	s := []string{"x"}
	s = appendUnique(s, "x")
	assert.Equal(t, []string{"x"}, s)
	s = appendUnique(s, "y")
	assert.Equal(t, []string{"x", "y"}, s)
}
