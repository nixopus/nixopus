package queue

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cryptossh "golang.org/x/crypto/ssh"
)

func TestExtractRequestIDFromChannel(t *testing.T) {
	assert.Equal(t, "abc", ExtractRequestIDFromChannel("machine:reply:abc"))
	assert.Empty(t, ExtractRequestIDFromChannel("other"))
}

func TestReplyMultiplexer_Dispatch_unknownWaiter(t *testing.T) {
	m := NewReplyMultiplexer()
	m.Dispatch("no-waiter", "payload")
}

func Test_getOrCreateProducerQueue_concurrent(t *testing.T) {
	s := miniredis.RunT(t)
	t.Cleanup(func() { s.Close() })
	opt, err := redis.ParseURL("redis://" + s.Addr())
	require.NoError(t, err)
	rdb := redis.NewClient(opt)
	t.Cleanup(func() { _ = rdb.Close() })
	Init(rdb)
	t.Cleanup(func() {
		_ = Close()
		redisClient = nil
	})

	const name = "concurrent-producer-q"
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_ = getOrCreateProducerQueue(name)
		}()
	}
	close(start)
	wg.Wait()
}

func Test_defaultMachineVerifySSHProbe_dialFailure(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	signer, err := cryptossh.ParsePrivateKey(pemBytes)
	require.NoError(t, err)
	cfg := &cryptossh.ClientConfig{
		User:            "root",
		Auth:            []cryptossh.AuthMethod{cryptossh.PublicKeys(signer)},
		HostKeyCallback: cryptossh.InsecureIgnoreHostKey(),
		Timeout:         200 * time.Millisecond,
	}
	err = defaultMachineVerifySSHProbe(context.Background(), "127.0.0.1:1", cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SSH dial failed")
}

func TestReplyMultiplexer_Start_ctxCancel(t *testing.T) {
	s := miniredis.RunT(t)
	t.Cleanup(func() { s.Close() })
	opt, err := redis.ParseURL("redis://" + s.Addr())
	require.NoError(t, err)
	rdb := redis.NewClient(opt)
	t.Cleanup(func() { _ = rdb.Close() })
	prev := redisClient
	redisClient = rdb
	t.Cleanup(func() { redisClient = prev })

	m := NewReplyMultiplexer()
	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx)
	cancel()
	time.Sleep(80 * time.Millisecond)
}

func TestReplyMultiplexer_Start_pubsubEmptyRequestID(t *testing.T) {
	s := miniredis.RunT(t)
	t.Cleanup(func() { s.Close() })
	opt, err := redis.ParseURL("redis://" + s.Addr())
	require.NoError(t, err)
	rdb := redis.NewClient(opt)
	t.Cleanup(func() { _ = rdb.Close() })
	prev := redisClient
	redisClient = rdb
	t.Cleanup(func() { redisClient = prev })

	m := NewReplyMultiplexer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.Start(ctx)
	time.Sleep(60 * time.Millisecond)
	require.NoError(t, rdb.Publish(ctx, "machine:reply:", "ignored").Err())
	time.Sleep(40 * time.Millisecond)
}
