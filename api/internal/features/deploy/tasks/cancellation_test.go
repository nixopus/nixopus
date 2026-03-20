package tasks

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func TestCancelledStatusExists(t *testing.T) {
	var s shared_types.Status = shared_types.Cancelled
	if s != "cancelled" {
		t.Errorf("expected Cancelled status to be 'cancelled', got %q", s)
	}
}

func TestRegisterAndCancelDeployment(t *testing.T) {
	svc := &TaskService{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deploymentID := "test-deployment-id"
	svc.RegisterCancellation(deploymentID, cancel)

	select {
	case <-ctx.Done():
		t.Fatal("context should not be cancelled yet")
	default:
	}

	err := svc.CancelDeployment(deploymentID)
	if err != nil {
		t.Fatalf("CancelDeployment returned unexpected error: %v", err)
	}

	select {
	case <-ctx.Done():
	default:
		t.Fatal("context should be cancelled after CancelDeployment")
	}
}

func TestCancelDeploymentNotFound(t *testing.T) {
	svc := &TaskService{}

	err := svc.CancelDeployment("non-existent-id")
	if err == nil {
		t.Fatal("expected error when cancelling non-existent deployment")
	}
}

func TestDeregisterCancellation(t *testing.T) {
	svc := &TaskService{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deploymentID := "test-deregister"
	svc.RegisterCancellation(deploymentID, cancel)
	svc.DeregisterCancellation(deploymentID)

	err := svc.CancelDeployment(deploymentID)
	if err == nil {
		t.Fatal("expected error after deregistering, but got nil")
	}

	select {
	case <-ctx.Done():
		t.Fatal("context should NOT be cancelled after deregister + cancel attempt")
	default:
	}
}

func TestConcurrentCancellations(t *testing.T) {
	svc := &TaskService{}
	const n = 100

	cancels := make([]context.CancelFunc, n)
	ctxs := make([]context.Context, n)

	for i := 0; i < n; i++ {
		ctxs[i], cancels[i] = context.WithCancel(context.Background())
		svc.RegisterCancellation(fmt.Sprintf("deployment-%d", i), cancels[i])
	}

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = svc.CancelDeployment(fmt.Sprintf("deployment-%d", idx))
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		select {
		case <-ctxs[i].Done():
		case <-time.After(time.Second):
			t.Errorf("context %d was not cancelled", i)
		}
	}
}

func TestCheckCancelledNotCancelled(t *testing.T) {
	ctx := context.Background()
	err := checkCancelled(ctx)
	if err != nil {
		t.Fatalf("expected nil for non-cancelled context, got: %v", err)
	}
}

func TestCheckCancelledIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := checkCancelled(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestRemoteBuildReaderContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	blockingReader := &blockingReader{ch: make(chan struct{})}
	closedCh := make(chan struct{})
	sess := &mockSession{closedCh: closedCh}

	reader := &remoteBuildReader{
		stdout:  blockingReader,
		session: sess,
		release: func() {},
		ctx:     ctx,
	}

	readDone := make(chan error, 1)
	go func() {
		buf := make([]byte, 128)
		_, err := reader.Read(buf)
		readDone <- err
	}()

	cancel()

	select {
	case <-closedCh:
	case <-time.After(2 * time.Second):
		t.Fatal("session was not closed after context cancellation")
	}

	close(blockingReader.ch)

	select {
	case err := <-readDone:
		if err == nil {
			t.Fatal("expected error from Read after cancel")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Read did not return after cancellation")
	}
}

type blockingReader struct {
	ch chan struct{}
}

func (r *blockingReader) Read(p []byte) (int, error) {
	<-r.ch
	return 0, context.Canceled
}

type mockSession struct {
	closedCh chan struct{}
	closed   bool
}

func (s *mockSession) Wait() error { return nil }
func (s *mockSession) Close() error {
	if !s.closed {
		s.closed = true
		close(s.closedCh)
	}
	return nil
}

func (s *mockSession) Signal(sig interface{}) error { return nil }
