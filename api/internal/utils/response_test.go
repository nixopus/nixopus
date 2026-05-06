package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendJSONResponse(t *testing.T) {
	t.Run("nil data", func(t *testing.T) {
		rr := httptest.NewRecorder()
		SendJSONResponse(rr, "ok", "m", nil)
		assert.Equal(t, http.StatusOK, rr.Code) // 200 default on recorder
		var got map[string]any
		require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
		assert.Equal(t, "ok", got["status"])
		assert.Equal(t, "m", got["message"])
	})

	t.Run("with data", func(t *testing.T) {
		rr := httptest.NewRecorder()
		SendJSONResponse(rr, "ok", "m", map[string]any{"x": 1})
		var got map[string]any
		require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
		assert.Equal(t, float64(1), got["data"].(map[string]any)["x"])
	})

	t.Run("marshal data fails", func(t *testing.T) {
		rr := httptest.NewRecorder()
		bad := map[string]any{"c": make(chan int)}
		SendJSONResponse(rr, "ok", "m", bad)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var got map[string]any
		require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
		assert.Equal(t, "Internal Server Error", got["title"])
		assert.Equal(t, float64(500), got["status"])
		assert.Equal(t, "failed to encode response data", got["detail"])
	})

	t.Run("encode to writer fails", func(t *testing.T) {
		w := &failingResponseWriter{header: make(http.Header)}
		SendJSONResponse(w, "ok", "m", map[string]any{"a": 1})
	})
}

type failingResponseWriter struct {
	header http.Header
}

func (f *failingResponseWriter) Header() http.Header { return f.header }
func (f *failingResponseWriter) WriteHeader(int)     {}
func (f *failingResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestSendErrorResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	SendErrorResponse(rr, "bad", http.StatusBadRequest)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	assert.Equal(t, "Bad Request", got["title"])
	assert.Equal(t, float64(400), got["status"])
	assert.Equal(t, "bad", got["detail"])
}

func TestSendErrorResponse_encodeFails(t *testing.T) {
	w := &failingResponseWriter{header: make(http.Header)}
	SendErrorResponse(w, "e", 400)
}
