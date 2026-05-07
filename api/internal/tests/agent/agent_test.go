//go:build integration

package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Thread CRUD ──────────────────────────────────────────────────────────────

func TestCreateThread_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /agent/threads without auth returns 401"),
		Post(tests.GetAgentThreadsURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestCreateThread(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /agent/threads with auth returns 200 and thread data"),
		Post(tests.GetAgentThreadsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestListThreads_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /agent/threads without auth returns 401"),
		Get(tests.GetAgentThreadsURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestListThreads(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// Create a thread first
	Test(t,
		Description("Create a thread before listing"),
		Post(tests.GetAgentThreadsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)

	Test(t,
		Description("GET /agent/threads returns success"),
		Get(tests.GetAgentThreadsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestDeleteThread(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	cookies := auth.GetAuthCookiesHeader()
	orgID := auth.OrganizationID

	// Create a thread
	req, _ := http.NewRequest("POST", tests.GetAgentThreadsURL(), nil)
	req.Header.Set("Cookie", cookies)
	req.Header.Set("X-Organization-ID", orgID)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	var createBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createBody)
	resp.Body.Close()

	data := createBody["data"].(map[string]interface{})
	threadID := data["id"].(string)

	Test(t,
		Description("DELETE /agent/threads/:id deletes the thread"),
		Delete(tests.GetAgentThreadURL(threadID)),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

// ─── Chat (blocking, real LLM) ──────────────────────────────────────────────

func TestChat_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /agent/chat without auth returns 401"),
		Post(tests.GetAgentChatURL()),
		Send().Body().JSON(map[string]interface{}{"input": "hello"}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestChat_EmptyInput(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /agent/chat with empty input returns 400"),
		Post(tests.GetAgentChatURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"input": ""}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestChat_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real LLM call in short mode")
	}

	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	cookies := auth.GetAuthCookiesHeader()
	orgID := auth.OrganizationID

	client := &http.Client{Timeout: 120 * time.Second}
	payload := `{"input": "Say hello in exactly 3 words"}`
	req, _ := http.NewRequest("POST", tests.GetAgentChatURL(), strings.NewReader(payload))
	req.Header.Set("Cookie", cookies)
	req.Header.Set("X-Organization-ID", orgID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.NotEmpty(t, body["thread_id"])
	assert.NotEmpty(t, body["response"])

	t.Logf("Agent response: %s", body["response"])
}

func TestChat_ModelOverrideInBody(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real LLM call in short mode")
	}

	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	cookies := auth.GetAuthCookiesHeader()
	orgID := auth.OrganizationID

	client := &http.Client{Timeout: 120 * time.Second}
	payload := `{"input": "Reply with just the word OK", "model": "openai/gpt-4o-mini"}`
	req, _ := http.NewRequest("POST", tests.GetAgentChatURL(), strings.NewReader(payload))
	req.Header.Set("Cookie", cookies)
	req.Header.Set("X-Organization-ID", orgID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.NotEmpty(t, body["response"])

	t.Logf("Model override response: %s", body["response"])
}

func TestChat_ModelOverrideViaHeader(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real LLM call in short mode")
	}

	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	cookies := auth.GetAuthCookiesHeader()
	orgID := auth.OrganizationID

	client := &http.Client{Timeout: 120 * time.Second}
	payload := `{"input": "Reply with just the word YES"}`
	req, _ := http.NewRequest("POST", tests.GetAgentChatURL(), strings.NewReader(payload))
	req.Header.Set("Cookie", cookies)
	req.Header.Set("X-Organization-ID", orgID)
	req.Header.Set("X-Model-Id", "openai/gpt-4o-mini")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.NotEmpty(t, body["response"])
}

func TestChat_ConversationMemory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real LLM call in short mode")
	}

	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	cookies := auth.GetAuthCookiesHeader()
	orgID := auth.OrganizationID
	client := &http.Client{Timeout: 120 * time.Second}

	// First message: ask it to remember something
	payload := `{"input": "Remember the number 42. Just confirm you remembered it."}`
	req, _ := http.NewRequest("POST", tests.GetAgentChatURL(), strings.NewReader(payload))
	req.Header.Set("Cookie", cookies)
	req.Header.Set("X-Organization-ID", orgID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	var firstResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&firstResp)
	resp.Body.Close()
	threadID := firstResp["thread_id"].(string)
	require.NotEmpty(t, threadID)

	// Second message on same thread: ask it to recall
	payload = fmt.Sprintf(`{"input": "What number did I ask you to remember?", "thread_id": "%s"}`, threadID)
	req, _ = http.NewRequest("POST", tests.GetAgentChatURL(), strings.NewReader(payload))
	req.Header.Set("Cookie", cookies)
	req.Header.Set("X-Organization-ID", orgID)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var secondResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&secondResp)
	assert.Equal(t, threadID, secondResp["thread_id"])
	assert.Contains(t, strings.ToLower(secondResp["response"].(string)), "42")

	t.Logf("Memory response: %s", secondResp["response"])
}

func TestChat_ThreadMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real LLM call in short mode")
	}

	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	cookies := auth.GetAuthCookiesHeader()
	orgID := auth.OrganizationID
	client := &http.Client{Timeout: 120 * time.Second}

	// Chat to create a thread with messages
	payload := `{"input": "Say hi"}`
	req, _ := http.NewRequest("POST", tests.GetAgentChatURL(), strings.NewReader(payload))
	req.Header.Set("Cookie", cookies)
	req.Header.Set("X-Organization-ID", orgID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	var chatResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&chatResp)
	resp.Body.Close()
	threadID := chatResp["thread_id"].(string)

	// Get messages for that thread
	Test(t,
		Description("GET /agent/threads/:id/messages returns stored messages"),
		Get(tests.GetAgentThreadMessagesURL(threadID)),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

// ─── Stream (SSE, real LLM) ─────────────────────────────────────────────────

func TestStreamChat_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /agent/chat/stream without auth returns 401"),
		Post(tests.GetAgentStreamURL()),
		Send().Body().JSON(map[string]interface{}{"input": "hello"}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestStreamChat_EmptyInput(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /agent/chat/stream with empty input returns 400"),
		Post(tests.GetAgentStreamURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"input": ""}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestStreamChat_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real LLM call in short mode")
	}

	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	cookies := auth.GetAuthCookiesHeader()
	orgID := auth.OrganizationID
	client := &http.Client{Timeout: 120 * time.Second}

	payload := `{"input": "Count from 1 to 3"}`
	req, _ := http.NewRequest("POST", tests.GetAgentStreamURL(), strings.NewReader(payload))
	req.Header.Set("Cookie", cookies)
	req.Header.Set("X-Organization-ID", orgID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/event-stream")

	events := readSSEEvents(t, resp.Body)
	require.NotEmpty(t, events)

	hasContent := false
	hasDone := false
	for _, ev := range events {
		if ev.event == "content" {
			hasContent = true
		}
		if ev.event == "done" {
			hasDone = true
		}
	}
	assert.True(t, hasContent, "expected at least one content event")
	assert.True(t, hasDone, "expected done event")

	t.Logf("Received %d SSE events", len(events))
}

func TestStreamChat_WithModelOverride(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real LLM call in short mode")
	}

	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	cookies := auth.GetAuthCookiesHeader()
	orgID := auth.OrganizationID
	client := &http.Client{Timeout: 120 * time.Second}

	payload := `{"input": "Say OK", "model": "openai/gpt-4o-mini"}`
	req, _ := http.NewRequest("POST", tests.GetAgentStreamURL(), strings.NewReader(payload))
	req.Header.Set("Cookie", cookies)
	req.Header.Set("X-Organization-ID", orgID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	events := readSSEEvents(t, resp.Body)
	hasDone := false
	for _, ev := range events {
		if ev.event == "done" {
			hasDone = true
			var data map[string]interface{}
			json.Unmarshal([]byte(ev.data), &data)
			assert.NotEmpty(t, data["thread_id"])
		}
	}
	assert.True(t, hasDone)
}

// ─── Cancel ───────────────────────────────────────────────────────────────────

func TestCancelStream_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /agent/chat/cancel without auth returns 401"),
		Post(tests.GetAgentCancelURL()),
		Send().Body().JSON(map[string]interface{}{"thread_id": "some-id"}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestCancelStream_MissingThreadID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /agent/chat/cancel with empty thread_id returns 400"),
		Post(tests.GetAgentCancelURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"thread_id": ""}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCancelStream_NotFound(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /agent/chat/cancel with non-existent thread_id returns not_found"),
		Post(tests.GetAgentCancelURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"thread_id": "nonexistent-thread-id"}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("not_found"),
	)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

type sseEvent struct {
	event string
	data  string
}

func readSSEEvents(t *testing.T, body io.Reader) []sseEvent {
	t.Helper()
	var events []sseEvent
	scanner := bufio.NewScanner(body)

	var currentEvent, currentData string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			currentData = strings.TrimPrefix(line, "data: ")
		} else if line == "" && currentEvent != "" {
			events = append(events, sseEvent{event: currentEvent, data: currentData})
			currentEvent = ""
			currentData = ""
		}
	}
	if currentEvent != "" {
		events = append(events, sseEvent{event: currentEvent, data: currentData})
	}
	return events
}
