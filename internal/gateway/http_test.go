package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"forgemate/internal/sidecar"
)

func TestHandleReadyReflectsState(t *testing.T) {
	s := NewHTTPServer(func() bool { return false }, sidecar.NewSupervisor())
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	resp := httptest.NewRecorder()

	s.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.Code)
	}
}

func TestHandleConnectValidateAcceptsValidFrame(t *testing.T) {
	s := NewHTTPServer(func() bool { return true }, sidecar.NewSupervisor())

	payload := map[string]any{
		"type":   "req",
		"method": "connect",
		"params": map[string]any{
			"protocol": 1,
			"role":     "operator",
			"client": map[string]any{
				"id": "admin-cli",
			},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/connect/validate", bytes.NewReader(raw))
	resp := httptest.NewRecorder()
	s.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestHandleConnectValidateRejectsInvalidFrame(t *testing.T) {
	s := NewHTTPServer(func() bool { return true }, sidecar.NewSupervisor())
	req := httptest.NewRequest(http.MethodPost, "/v1/connect/validate", bytes.NewReader([]byte(`{"type":"event"}`)))
	resp := httptest.NewRecorder()

	s.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}
