package redisclient

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_invalidURL(t *testing.T) {
	c, err := New("://bad")
	require.Error(t, err)
	assert.Nil(t, c)

	c, err = New("")
	require.Error(t, err)
	assert.Nil(t, c)
}

func TestNew_appliesTimeoutsAndPool(t *testing.T) {
	c, err := New("redis://127.0.0.1:6379/0")
	require.NoError(t, err)
	require.NotNil(t, c)
	t.Cleanup(func() { _ = c.Close() })

	opt := c.Options()
	assert.Equal(t, 10, opt.MinIdleConns)
	assert.Equal(t, 5*time.Second, opt.DialTimeout)
	assert.Equal(t, 3*time.Second, opt.ReadTimeout)
	assert.Equal(t, 3*time.Second, opt.WriteTimeout)
}
