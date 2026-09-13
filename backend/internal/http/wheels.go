package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"spinwheel/backend/internal/repository"
	"spinwheel/backend/pkg/models"
)

// wheelItemIn is one item in a create/sync body. ID is optional on sync
// (empty = new item, server mints the ID); weight <= 0 defaults to 1.0 and
// color defaults to "" — the FE sends option-only items, and the same
// defaults must be assumed when it hashes the list client-side.
type wheelItemIn struct {
	ID     string  `json:"id"`
	Option string  `json:"option"`
	Color  string  `json:"color"`
	Weight float64 `json:"weight"`
}

type createWheelRequest struct {
	Name  string        `json:"name"`
	Items []wheelItemIn `json:"items"`
}

type syncWheelRequest struct {
	Name  string        `json:"name"`
	Items []wheelItemIn `json:"items"`
}

// wheelJSON mirrors the stored wheel for create/sync responses.
type wheelJSON struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Items []itemJSON `json:"items"`
}

func toWheelJSON(wheel *models.Wheel) wheelJSON {
	items := make([]itemJSON, 0, len(wheel.Items))
	for _, it := range wheel.Items {
		items = append(items, itemJSON{
			ID:     it.ID,
			Option: it.Option,
			Color:  it.Color,
			Weight: it.Weight,
		})
	}
	return wheelJSON{ID: wheel.ID, Name: wheel.Name, Items: items}
}

// normalizeItems trims options, rejects blanks, and applies the weight
// default. It returns the model items the service persists.
func normalizeItems(in []wheelItemIn) ([]models.WheelItem, error) {
	items := make([]models.WheelItem, 0, len(in))
	for _, it := range in {
		option := strings.TrimSpace(it.Option)
		if option == "" {
			return nil, errors.New("item option must not be empty")
		}
		weight := it.Weight
		if weight <= 0 {
			weight = 1.0
		}
		items = append(items, models.WheelItem{
			ID:     strings.TrimSpace(it.ID),
			Option: option,
			Color:  it.Color,
			Weight: weight,
		})
	}
	return items, nil
}

func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if r.ContentLength > maxBodyBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "body too large")
		return false
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

// handleCreateWheel serves POST /v1/wheels.
func (s *Server) handleCreateWheel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req createWheelRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	items, err := normalizeItems(req.Items)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	wheel, err := s.svc.CreateWheel(&models.Wheel{
		Name:  strings.TrimSpace(req.Name),
		Items: items,
	})
	if err != nil {
		s.log.Error("create wheel", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, toWheelJSON(wheel))
}

// handleSyncWheel serves PUT /v1/wheels/{id}: full-list replace. Known IDs
// are preserved, unknown/empty IDs are minted, dropped IDs are deleted.
func (s *Server) handleSyncWheel(w http.ResponseWriter, r *http.Request, wheelID string) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req syncWheelRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	items, err := normalizeItems(req.Items)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	wheel, err := s.svc.UpdateWheel(&models.Wheel{
		ID:    wheelID,
		Name:  strings.TrimSpace(req.Name),
		Items: items,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "wheel not found")
			return
		}
		s.log.Error("sync wheel", zap.String("wheel_id", wheelID), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, toWheelJSON(wheel))
}
