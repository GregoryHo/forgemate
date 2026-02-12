package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	FrameTypeReq   = "req"
	FrameTypeRes   = "res"
	FrameTypeEvent = "event"
	FrameTypeError = "error"
)

// Frame is the control-plane envelope for WS-RPC style transport.
type Frame struct {
	Type   string          `json:"type"`
	ID     string          `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
}

// ConnectParams is the required first request payload.
type ConnectParams struct {
	Protocol int    `json:"protocol"`
	Role     string `json:"role"`
	Token    string `json:"token,omitempty"`
	Client   struct {
		ID      string `json:"id"`
		Version string `json:"version,omitempty"`
	} `json:"client"`
}

func DecodeFrame(raw []byte) (Frame, error) {
	var frame Frame
	if err := json.Unmarshal(raw, &frame); err != nil {
		return Frame{}, fmt.Errorf("decode frame: %w", err)
	}
	if frame.Type == "" {
		return Frame{}, errors.New("missing frame type")
	}
	return frame, nil
}

// ValidateConnectFirst enforces the connect-first handshake discipline.
func ValidateConnectFirst(frame Frame) error {
	if frame.Type != FrameTypeReq {
		return fmt.Errorf("first frame must be req, got %q", frame.Type)
	}
	if frame.Method != "connect" {
		return fmt.Errorf("first method must be connect, got %q", frame.Method)
	}
	var params ConnectParams
	if err := json.Unmarshal(frame.Params, &params); err != nil {
		return fmt.Errorf("invalid connect params: %w", err)
	}
	if params.Protocol <= 0 {
		return errors.New("protocol must be positive")
	}
	switch strings.TrimSpace(params.Role) {
	case "operator", "app", "node":
	default:
		return fmt.Errorf("invalid role %q", params.Role)
	}
	if strings.TrimSpace(params.Client.ID) == "" {
		return errors.New("client.id is required")
	}
	return nil
}
