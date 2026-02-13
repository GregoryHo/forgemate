package gateway

import (
	"encoding/json"
	"testing"
)

func mustJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return raw
}

func TestValidateConnectFirstAcceptsValidFrame(t *testing.T) {
	frame := Frame{
		Type:   FrameTypeReq,
		Method: "connect",
		Params: mustJSON(t, map[string]any{
			"protocol": 1,
			"role":     "operator",
			"client": map[string]any{
				"id":      "admin-ui",
				"version": "0.1.0",
			},
		}),
	}

	if err := ValidateConnectFirst(frame); err != nil {
		t.Fatalf("expected valid frame: %v", err)
	}
}

func TestValidateConnectFirstRejectsInvalidRole(t *testing.T) {
	frame := Frame{
		Type:   FrameTypeReq,
		Method: "connect",
		Params: mustJSON(t, map[string]any{
			"protocol": 1,
			"role":     "guest",
			"client": map[string]any{
				"id": "client-a",
			},
		}),
	}

	if err := ValidateConnectFirst(frame); err == nil {
		t.Fatal("expected role validation error")
	}
}

func TestDecodeFrameRejectsMissingType(t *testing.T) {
	_, err := DecodeFrame([]byte(`{"id":"1","method":"connect"}`))
	if err == nil {
		t.Fatal("expected missing frame type error")
	}
}
