package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"spinwheel/backend/internal/repository"
	"spinwheel/backend/internal/service"
)

// spinRequest is POST /v1/wheels/{id}/spin {client_seed, items_hash}.
// session_id is an optional client echo; the sw_sid cookie wins when valid.
type spinRequest struct {
	ClientSeed string `json:"client_seed"`
	ItemsHash  string `json:"items_hash"`
	SessionID  string `json:"session_id"`
}

type itemJSON struct {
	ID     string  `json:"id"`
	Option string  `json:"option"`
	Color  string  `json:"color"`
	Weight float64 `json:"weight"`
}

// spinResponse mirrors the gRPC SpinWheelResponse flat fields.
type spinResponse struct {
	SpinID      string   `json:"spin_id"`
	WinnerIndex int      `json:"winner_index"`
	WinnerItem  itemJSON `json:"winner_item"`
	Nonce       string   `json:"nonce"`
	Sig         string   `json:"sig"`
	ExpiresAt   string   `json:"expires_at"`
}

func (s *Server) handleSpin(w http.ResponseWriter, r *http.Request, wheelID string) {
	if r.ContentLength > maxBodyBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "body too large")
		return
	}
	var req spinRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	sessionID := s.ensureSessionID(w, r, req.SessionID)

	wheel, err := s.svc.GetWheel(wheelID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "wheel not found")
			return
		}
		s.log.Error("spin get wheel", zap.String("wheel_id", wheelID), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Strict freshness check, mirroring the gRPC handler: a spin drawn
	// against an unknown item list could never verify.
	if req.ItemsHash == "" {
		writeError(w, http.StatusBadRequest, "wheel_changed: items_hash is required")
		return
	}
	if want := service.ComputeItemsHash(wheel.Items); want != req.ItemsHash {
		writeError(w, http.StatusBadRequest, "wheel_changed: items_hash mismatch")
		return
	}

	ip := clientIP(r)
	result, err := s.svc.SpinWheelSecure(r.Context(), wheelID, service.SpinOptions{
		ClientSeed: req.ClientSeed,
		ItemsHash:  req.ItemsHash,
		SessionID:  sessionID,
		IPHash:     ipHash(ip),
		UA:         r.UserAgent(),
		CFRay:      cfRay(r),
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "wheel not found")
			return
		}
		s.log.Error("spin secure", zap.String("wheel_id", wheelID), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	s.log.Info("spin",
		zap.String("spin_id", result.SpinID),
		zap.String("wheel_id", wheelID),
		zap.String("cf_ray", cfRay(r)),
	)

	writeJSON(w, http.StatusOK, spinResponse{
		SpinID:      result.SpinID,
		WinnerIndex: result.WinnerIdx,
		WinnerItem: itemJSON{
			ID:     result.Item.ID,
			Option: result.Item.Option,
			Color:  result.Item.Color,
			Weight: result.Item.Weight,
		},
		Nonce:     result.Nonce,
		Sig:       result.Sig,
		ExpiresAt: result.ExpiresAt.UTC().Format(time.RFC3339),
	})
}
