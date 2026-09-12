package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"spinwheel/backend/pkg/models"
)

// eventRequest is POST /v1/events. Session comes from the sw_sid cookie;
// session_id is an optional client echo used when the cookie is absent.
type eventRequest struct {
	Type      string `json:"type"`
	WheelID   string `json:"wheel_id"`
	SpinID    string `json:"spin_id"`
	Path      string `json:"path"`
	Sig       string `json:"sig"`
	SessionID string `json:"session_id"`
	ClientTS  string `json:"client_ts"`
}

// handleEvents records one analytics event and returns 204. Recording is
// idempotent on (session_id, client_ts, type, spin_id); redelivery is a
// no-op, not an error.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if r.ContentLength > maxBodyBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "body too large")
		return
	}
	var req eventRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var eventType string
	switch strings.ToUpper(strings.TrimSpace(req.Type)) {
	case "PAGEVIEW":
		eventType = "PAGEVIEW"
	case "SPIN_START":
		eventType = "SPIN_START"
	case "SPIN_END":
		eventType = "SPIN_END"
	default:
		writeError(w, http.StatusBadRequest, "event type is required")
		return
	}

	if req.WheelID != "" {
		if _, err := uuid.Parse(req.WheelID); err != nil {
			writeError(w, http.StatusBadRequest, "invalid wheel_id")
			return
		}
	}
	if req.SpinID != "" {
		if _, err := uuid.Parse(req.SpinID); err != nil {
			writeError(w, http.StatusBadRequest, "invalid spin_id")
			return
		}
	}

	var clientTS time.Time
	if strings.TrimSpace(req.ClientTS) != "" {
		var err error
		clientTS, err = time.Parse(time.RFC3339, strings.TrimSpace(req.ClientTS))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid client_ts (want RFC3339)")
			return
		}
	}

	sessionID := s.ensureSessionID(w, r, req.SessionID)

	if _, err := s.svc.RecordEvent(r.Context(), models.EventParams{
		SessionID: sessionID,
		Type:      eventType,
		WheelID:   req.WheelID,
		SpinID:    req.SpinID,
		Path:      req.Path,
		Sig:       req.Sig,
		ClientTS:  clientTS,
		IPHash:    ipHash(clientIP(r)),
		UA:        r.UserAgent(),
		CFRay:     cfRay(r),
	}); err != nil {
		s.log.Error("record event",
			zap.String("type", eventType),
			zap.String("cf_ray", cfRay(r)),
			zap.Error(err),
		)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
