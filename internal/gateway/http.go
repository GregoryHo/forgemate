package gateway

import (
	"encoding/json"
	"net/http"
	"time"

	"forgemate/internal/sidecar"
)

// HTTPServer exposes minimal operational and protocol validation endpoints.
type HTTPServer struct {
	mux       *http.ServeMux
	stateOK   func() bool
	sidecar   *sidecar.Supervisor
	startedAt time.Time
}

func NewHTTPServer(stateOK func() bool, sidecarSupervisor *sidecar.Supervisor) *HTTPServer {
	s := &HTTPServer{
		mux:       http.NewServeMux(),
		stateOK:   stateOK,
		sidecar:   sidecarSupervisor,
		startedAt: time.Now().UTC(),
	}
	s.routes()
	return s
}

func (s *HTTPServer) Handler() http.Handler {
	return s.mux
}

func (s *HTTPServer) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/ready", s.handleReady)
	s.mux.HandleFunc("/v1/connect/validate", s.handleConnectValidate)
}

func (s *HTTPServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":        true,
		"service":   "forgemate-gateway",
		"startedAt": s.startedAt,
		"sidecar":   s.sidecar.Status(),
	})
}

func (s *HTTPServer) handleReady(w http.ResponseWriter, _ *http.Request) {
	ready := s.stateOK()
	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":    ready,
		"ready": ready,
	})
}

func (s *HTTPServer) handleConnectValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var frame Frame
	if err := json.NewDecoder(r.Body).Decode(&frame); err != nil {
		http.Error(w, "invalid frame JSON", http.StatusBadRequest)
		return
	}

	if err := ValidateConnectFirst(frame); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"protocol": 1,
		"message":  "connect accepted",
	})
}
