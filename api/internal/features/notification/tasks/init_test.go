package tasks

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification/channel"
	"github.com/nixopus/nixopus/api/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubChannel struct {
	typeName string
	sendErr  error
	lastMsg  channel.Message
}

func (s *stubChannel) Type() string { return s.typeName }

func (s *stubChannel) Send(ctx context.Context, msg channel.Message) error {
	_ = ctx
	s.lastMsg = msg
	return s.sendErr
}

func TestEnqueue_NotInitialized(t *testing.T) {
	prevQ, prevT := NotificationQueue, TaskSendNotification
	NotificationQueue, TaskSendNotification = nil, nil
	t.Cleanup(func() {
		NotificationQueue, TaskSendNotification = prevQ, prevT
	})

	err := Enqueue(channel.DeliveryPayload{Channel: "email"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "notification queue not initialized")
}

func TestSendNotificationTaskHandler_UnknownChannel(t *testing.T) {
	h := sendNotificationTaskHandler(map[string]channel.Channel{}, logger.NewLogger())
	err := h(context.Background(), channel.DeliveryPayload{
		Channel: "missing",
		Message: channel.Message{To: "x", Body: "b"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown notification channel: missing")
}

func TestSendNotificationTaskHandler_SendError(t *testing.T) {
	sendErr := errors.New("transport down")
	stub := &stubChannel{typeName: "email", sendErr: sendErr}
	h := sendNotificationTaskHandler(map[string]channel.Channel{"email": stub}, logger.NewLogger())

	err := h(context.Background(), channel.DeliveryPayload{
		Channel: "email",
		Message: channel.Message{To: "u@x.com", Body: "hello"},
	})
	require.ErrorIs(t, err, sendErr)
}

func TestSendNotificationTaskHandler_Success(t *testing.T) {
	stub := &stubChannel{typeName: "slack"}
	h := sendNotificationTaskHandler(map[string]channel.Channel{"slack": stub}, logger.NewLogger())

	err := h(context.Background(), channel.DeliveryPayload{
		Channel: "slack",
		Message: channel.Message{To: "#alerts", Body: "ok"},
	})
	require.NoError(t, err)
	assert.Equal(t, "#alerts", stub.lastMsg.To)
	assert.Equal(t, "ok", stub.lastMsg.Body)
}

func TestSetupNotificationQueue_And_Enqueue(t *testing.T) {
	mr := miniredis.RunT(t)
	opt, err := redis.ParseURL("redis://" + mr.Addr())
	require.NoError(t, err)
	rdb := redis.NewClient(opt)
	t.Cleanup(func() { _ = rdb.Close() })

	queue.Init(rdb)
	t.Cleanup(func() { _ = queue.Close() })

	stub := &stubChannel{typeName: "email"}
	SetupNotificationQueue(map[string]channel.Channel{"email": stub}, logger.NewLogger())

	require.NotNil(t, NotificationQueue)
	require.NotNil(t, TaskSendNotification)

	err = Enqueue(channel.DeliveryPayload{
		Channel:        "email",
		OrganizationID: "org-1",
		UserID:         "user-1",
		Message:        channel.Message{To: "a@b.com", Body: "queued"},
	})
	require.NoError(t, err)
}
