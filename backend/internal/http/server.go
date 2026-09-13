// Package http serves the public browser-facing JSON API. gRPC stays
// internal; this gateway translates HTTP/JSON + cookies into service
// calls and maps errors to HTTP statuses.
//
// Routes (contract v1):
//
//	GET  /healthz                -> 200 {"status":"ok"}
//	POST /v1/wheels/{id}/spin    -> verifiable spin decision JSON
//	POST /v1/events              -> 204 (analytics, idempotent)
//
// Deliberately stdlib-only (no chi/grpc-gateway): three routes do not
// justify the dependency + codegen churn. Routing uses subtree patterns
// only, compatible with the go 1.21 language version.
package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"spinwheel/backend/internal/middleware"
	"spinwheel/backend/internal/service"
)

// maxBodyBytes caps JSON request bodies (contract: 1MB).
const maxBodyBytes = 1 << 20

// Config wires the gateway. Service is required; everything else has a
// safe zero value (nop logger, no CORS origins, plain cookies, no limit).
type Config struct {
	Service service.WheelService
	Logger  *zap.Logger
	// CORSOrigins is the allowlist echoed in Access-Control-Allow-Origin.
	// Empty denies all cross-origin reads (no header emitted).
	CORSOrigins []string
	// SecureCookies sets Secure on sw_sid (production HTTPS).
	SecureCookies bool
	// RequestsPerMin is the per IP+session rate limit; <=0 disables it.
	RequestsPerMin int
}

// Server is the HTTP gateway. Construct with NewServer; serve Handler().
type Server struct {
	svc           service.WheelService
	log           *zap.Logger
	cors          map[string]struct{}
	secureCookies bool
	limiter       *middleware.Limiter
}

// NewServer builds the gateway from Config.
func NewServer(cfg Config) *Server {
	log := cfg.Logger
	if log == nil {
		log = zap.NewNop()
	}
	s := &Server{
		svc:           cfg.Service,
		log:           log,
		cors:          make(map[string]struct{}),
		secureCookies: cfg.SecureCookies,
	}
	for _, o := range cfg.CORSOrigins {
		if o = strings.TrimSpace(o); o != "" {
			s.cors[strings.ToLower(o)] = struct{}{}
		}
	}
	if cfg.RequestsPerMin > 0 {
		s.limiter = middleware.NewLimiter(cfg.RequestsPerMin, time.Minute)
	}
	return s
}

// Handler assembles the route tree and middleware. Outermost first:
// CORS (preflight short-circuits before limit/logging), rate limit,
// request logging, then the mux.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/v1/wheels", s.handleCreateWheel)
	mux.HandleFunc("/v1/wheels/", s.handleWheels)
	mux.HandleFunc("/v1/events", s.handleEvents)

	var h http.Handler = mux
	h = s.withLogging(h)
	if s.limiter != nil {
		h = s.withRateLimit(h)
	}
	h = s.withCORS(h)
	return h
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleWheels dispatches the /v1/wheels/ subtree: PUT /v1/wheels/{id}
// syncs the item list, POST /v1/wheels/{id}/spin draws.
func (s *Server) handleWheels(w http.ResponseWriter, r *http.Request) {
	rest, ok := strings.CutPrefix(r.URL.Path, "/v1/wheels/")
	if !ok || rest == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	id, action, found := strings.Cut(rest, "/")
	if !found {
		// Bare /v1/wheels/{id}: sync only.
		s.handleSyncWheel(w, r, id)
		return
	}
	if action != "spin" || id == "" || strings.Contains(action, "/") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	s.handleSpin(w, r, id)
}

// errorEnvelope is the JSON error body.
type errorEnvelope struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, errorEnvelope{Error: msg})
}
