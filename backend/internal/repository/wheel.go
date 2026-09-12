package repository

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	crand "crypto/rand"

	"github.com/google/uuid"

	"spinwheel/backend/pkg/models"
)

var (
	ErrNotFound = errors.New("not found")
)

// defaultWeight applies when an item's weight is <= 0 (per contract v1).
const defaultWeight = 1.0

// spinHistoryLimit bounds ListSpins responses.
const spinHistoryLimit = 50

const maxSpinHistoryLimit = 200

type WheelRepository interface {
	CreateWheel(wheel *models.Wheel) (*models.Wheel, error)
	GetWheel(id string) (*models.Wheel, error)
	ListWheels() ([]*models.Wheel, error)
	UpdateWheel(wheel *models.Wheel) (*models.Wheel, error)
	DeleteWheel(id string) error
	AddItem(wheelID string, item *models.WheelItem) (*models.WheelItem, error)
	UpdateItem(wheelID string, item *models.WheelItem) (*models.WheelItem, error)
	DeleteItem(wheelID string, itemID string) error
	// SpinWheel picks a winner with a weighted crypto/rand draw.
	// It does NOT persist: use RecordSpin when the decision must be audited
	// (the service layer calls RecordSpin, then attaches the HMAC sig).
	SpinWheel(id string) (*models.WheelItem, error)
	// RecordSpin draws a winner and persists the decision audit row.
	RecordSpin(ctx context.Context, params models.SpinParams) (*models.SpinRecord, error)
	// UpdateSpinSig attaches the service-computed HMAC to a stored spin.
	UpdateSpinSig(ctx context.Context, spinID string, sig string) error
	// ListSpins returns persisted spins newest-first, capped at limit
	// (default spinHistoryLimit, max maxSpinHistoryLimit).
	ListSpins(ctx context.Context, wheelID string, limit int) ([]*models.SpinRecord, error)
	// RecordEvent persists one analytics event; redelivery of the same
	// (session, type, client_ts, spin) is a no-op returning the stored row.
	RecordEvent(ctx context.Context, params models.EventParams) (*models.EventRecord, error)
}

type InMemoryWheelRepository struct {
	wheels map[string]*models.Wheel
	spins  map[string]*models.SpinRecord
	events map[string]*models.EventRecord
	mutex  sync.RWMutex
}

func NewInMemoryWheelRepository() *InMemoryWheelRepository {
	return &InMemoryWheelRepository{
		wheels: make(map[string]*models.Wheel),
		spins:  make(map[string]*models.SpinRecord),
		events: make(map[string]*models.EventRecord),
	}
}

func (r *InMemoryWheelRepository) CreateWheel(wheel *models.Wheel) (*models.Wheel, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	wheel.ID = newID()
	for i := range wheel.Items {
		wheel.Items[i].ID = newID()
	}
	r.wheels[wheel.ID] = wheel
	return wheel, nil
}

func (r *InMemoryWheelRepository) GetWheel(id string) (*models.Wheel, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	wheel, ok := r.wheels[id]
	if !ok {
		return nil, ErrNotFound
	}
	return wheel, nil
}

func (r *InMemoryWheelRepository) ListWheels() ([]*models.Wheel, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var wheels []*models.Wheel
	for _, wheel := range r.wheels {
		wheels = append(wheels, wheel)
	}
	return wheels, nil
}

func (r *InMemoryWheelRepository) UpdateWheel(wheel *models.Wheel) (*models.Wheel, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.wheels[wheel.ID]; !ok {
		return nil, ErrNotFound
	}

	for i := range wheel.Items {
		if wheel.Items[i].ID == "" {
			wheel.Items[i].ID = newID()
		}
	}

	r.wheels[wheel.ID] = wheel
	return wheel, nil
}

func (r *InMemoryWheelRepository) DeleteWheel(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.wheels[id]; !ok {
		return ErrNotFound
	}

	delete(r.wheels, id)
	return nil
}

func (r *InMemoryWheelRepository) AddItem(wheelID string, item *models.WheelItem) (*models.WheelItem, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	wheel, ok := r.wheels[wheelID]
	if !ok {
		return nil, ErrNotFound
	}

	item.ID = newID()
	wheel.Items = append(wheel.Items, *item)
	return item, nil
}

func (r *InMemoryWheelRepository) UpdateItem(wheelID string, item *models.WheelItem) (*models.WheelItem, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	wheel, ok := r.wheels[wheelID]
	if !ok {
		return nil, ErrNotFound
	}

	for i := range wheel.Items {
		if wheel.Items[i].ID == item.ID {
			wheel.Items[i].Option = item.Option
			wheel.Items[i].Color = item.Color
			wheel.Items[i].Weight = item.Weight
			return &wheel.Items[i], nil
		}
	}
	return nil, ErrNotFound
}

func (r *InMemoryWheelRepository) DeleteItem(wheelID string, itemID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	wheel, ok := r.wheels[wheelID]
	if !ok {
		return ErrNotFound
	}

	for i := range wheel.Items {
		if wheel.Items[i].ID == itemID {
			wheel.Items = append(wheel.Items[:i], wheel.Items[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (r *InMemoryWheelRepository) SpinWheel(id string) (*models.WheelItem, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	wheel, ok := r.wheels[id]
	if !ok {
		return nil, ErrNotFound
	}

	idx, err := weightedWinnerIndex(wheel.Items)
	if err != nil {
		return nil, err
	}
	winningItem := wheel.Items[idx]
	return &winningItem, nil
}

func (r *InMemoryWheelRepository) RecordSpin(ctx context.Context, params models.SpinParams) (*models.SpinRecord, error) {
	_ = ctx

	r.mutex.Lock()
	defer r.mutex.Unlock()

	wheel, ok := r.wheels[params.WheelID]
	if !ok {
		return nil, ErrNotFound
	}

	idx, err := weightedWinnerIndex(wheel.Items)
	if err != nil {
		return nil, err
	}

	seed, err := randomBytes(32)
	if err != nil {
		return nil, err
	}
	nonce, err := randomHex(16)
	if err != nil {
		return nil, err
	}

	expiresAt := params.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(60 * time.Second)
	}

	record := &models.SpinRecord{
		ID:           newID(),
		WheelID:      params.WheelID,
		WinnerIdx:    idx,
		WinnerItemID: wheel.Items[idx].ID,
		SeedServer:   seed,
		SeedClient:   params.SeedClient,
		ItemsHash:    params.ItemsHash,
		Nonce:        nonce,
		SessionID:    params.SessionID,
		IPHash:       params.IPHash,
		UA:           params.UA,
		CFRay:        params.CFRay,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}
	r.spins[record.ID] = record
	return record, nil
}

func (r *InMemoryWheelRepository) UpdateSpinSig(ctx context.Context, spinID string, sig string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	spin, ok := r.spins[spinID]
	if !ok {
		return ErrNotFound
	}
	spin.Sig = sig
	return nil
}

func (r *InMemoryWheelRepository) ListSpins(ctx context.Context, wheelID string, limit int) ([]*models.SpinRecord, error) {
	_ = ctx

	limit = normalizeLimit(limit)

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if _, ok := r.wheels[wheelID]; !ok {
		return nil, ErrNotFound
	}

	var out []*models.SpinRecord
	for _, spin := range r.spins {
		if spin.WheelID == wheelID {
			out = append(out, spin)
		}
	}
	// Newest first (insertion sort is fine for a test/dev store).
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].CreatedAt.After(out[i].CreatedAt) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *InMemoryWheelRepository) RecordEvent(ctx context.Context, params models.EventParams) (*models.EventRecord, error) {

	clientTS := params.ClientTS
	if clientTS.IsZero() {
		clientTS = time.Now()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	key := eventKey(params.SessionID, params.Type, clientTS, params.SpinID)
	if existing, ok := r.events[key]; ok {
		return existing, nil
	}

	record := &models.EventRecord{
		ID:        newID(),
		SessionID: params.SessionID,
		Type:      params.Type,
		WheelID:   params.WheelID,
		SpinID:    params.SpinID,
		Path:      params.Path,
		Sig:       params.Sig,
		ClientTS:  clientTS,
		IPHash:    params.IPHash,
		UA:        params.UA,
		CFRay:     params.CFRay,
		CreatedAt: time.Now(),
	}
	r.events[key] = record
	return record, nil
}

func eventKey(sessionID, eventType string, clientTS time.Time, spinID string) string {
	return sessionID + "|" + eventType + "|" + clientTS.UTC().Format(time.RFC3339Nano) + "|" + spinID
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return spinHistoryLimit
	}
	if limit > maxSpinHistoryLimit {
		return maxSpinHistoryLimit
	}
	return limit
}

func newID() string {
	return uuid.NewString()
}

// weightedWinnerIndex draws one index proportional to item weights
// (weight <= 0 counts as defaultWeight) using crypto/rand.
func weightedWinnerIndex(items []models.WheelItem) (int, error) {
	if len(items) == 0 {
		return 0, errors.New("wheel has no items")
	}

	total := 0.0
	weights := make([]float64, len(items))
	for i, item := range items {
		w := item.Weight
		if w <= 0 {
			w = defaultWeight
		}
		weights[i] = w
		total += w
	}

	f, err := randFloat64()
	if err != nil {
		return 0, err
	}

	target := f * total
	acc := 0.0
	for i, w := range weights {
		acc += w
		if target < acc {
			return i, nil
		}
	}
	return len(items) - 1, nil
}

// randFloat64 returns a uniform float64 in [0, 1) from crypto/rand.
func randFloat64() (float64, error) {
	var b [8]byte
	if _, err := crand.Read(b[:]); err != nil {
		return 0, err
	}
	u := binary.BigEndian.Uint64(b[:]) >> 11 // keep 53 bits
	return float64(u) / (1 << 53), nil
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := crand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func randomHex(n int) (string, error) {
	b, err := randomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
