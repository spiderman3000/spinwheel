package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"spinwheel/backend/internal/repository"
	"spinwheel/backend/pkg/models"
)

type WheelService interface {
	CreateWheel(wheel *models.Wheel) (*models.Wheel, error)
	GetWheel(id string) (*models.Wheel, error)
	ListWheels() ([]*models.Wheel, error)
	UpdateWheel(wheel *models.Wheel) (*models.Wheel, error)
	DeleteWheel(id string) error
	AddItem(wheelID string, item *models.WheelItem) (*models.WheelItem, error)
	UpdateItem(wheelID string, item *models.WheelItem) (*models.WheelItem, error)
	DeleteItem(wheelID string, itemID string) error
	// SpinWheel is the legacy read-only pick (no persistence, no sig).
	// The handler still uses it until step 3; new callers use SpinWheelSecure.
	SpinWheel(id string) (*models.WheelItem, error)
	// SpinWheelSecure draws via the repository (persisted audit row),
	// attaches the HMAC sig, and returns the verifiable decision.
	SpinWheelSecure(ctx context.Context, wheelID string, opts SpinOptions) (*SpinResult, error)
	// VerifySpin recomputes the HMAC over the claimed decision fields.
	VerifySpin(spinID, wheelID string, winnerIdx int, itemsHash, nonce, sig string) bool
	ListSpins(ctx context.Context, wheelID string, limit int) ([]*models.SpinRecord, error)
	RecordEvent(ctx context.Context, params models.EventParams) (*models.EventRecord, error)
}

// SpinOptions carries the caller-supplied spin inputs. SessionID/IPHash/UA/
// CFRay come from the HTTP gateway (never the client body); ClientSeed and
// ItemsHash come from the client per contract v1.
type SpinOptions struct {
	ClientSeed string
	ItemsHash  string
	SessionID  string
	IPHash     string
	UA         string
	CFRay      string
}

// SpinResult is the verifiable server decision returned to the client.
type SpinResult struct {
	SpinID    string
	WinnerIdx int
	Item      *models.WheelItem
	Nonce     string
	Sig       string
	ExpiresAt time.Time
}

type wheelService struct {
	repo       repository.WheelRepository
	hmacSecret []byte
}

func NewWheelService(repo repository.WheelRepository, hmacSecret string) WheelService {
	return &wheelService{
		repo:       repo,
		hmacSecret: []byte(hmacSecret),
	}
}

func (s *wheelService) CreateWheel(wheel *models.Wheel) (*models.Wheel, error) {
	return s.repo.CreateWheel(wheel)
}

func (s *wheelService) GetWheel(id string) (*models.Wheel, error) {
	return s.repo.GetWheel(id)
}

func (s *wheelService) ListWheels() ([]*models.Wheel, error) {
	return s.repo.ListWheels()
}

func (s *wheelService) UpdateWheel(wheel *models.Wheel) (*models.Wheel, error) {
	return s.repo.UpdateWheel(wheel)
}

func (s *wheelService) DeleteWheel(id string) error {
	return s.repo.DeleteWheel(id)
}

func (s *wheelService) AddItem(wheelID string, item *models.WheelItem) (*models.WheelItem, error) {
	return s.repo.AddItem(wheelID, item)
}

func (s *wheelService) UpdateItem(wheelID string, item *models.WheelItem) (*models.WheelItem, error) {
	return s.repo.UpdateItem(wheelID, item)
}

func (s *wheelService) DeleteItem(wheelID string, itemID string) error {
	return s.repo.DeleteItem(wheelID, itemID)
}

func (s *wheelService) SpinWheel(id string) (*models.WheelItem, error) {
	return s.repo.SpinWheel(id)
}

func (s *wheelService) SpinWheelSecure(ctx context.Context, wheelID string, opts SpinOptions) (*SpinResult, error) {
	if len(s.hmacSecret) == 0 {
		return nil, errors.New("HMAC secret not configured")
	}

	record, err := s.repo.RecordSpin(ctx, models.SpinParams{
		WheelID:    wheelID,
		SeedClient: opts.ClientSeed,
		ItemsHash:  opts.ItemsHash,
		SessionID:  opts.SessionID,
		IPHash:     opts.IPHash,
		UA:         opts.UA,
		CFRay:      opts.CFRay,
	})
	if err != nil {
		return nil, err
	}

	sig := signSpin(s.hmacSecret, record.ID, wheelID, record.WinnerIdx, opts.ItemsHash, record.Nonce)
	if err := s.repo.UpdateSpinSig(ctx, record.ID, sig); err != nil {
		return nil, err
	}

	wheel, err := s.repo.GetWheel(wheelID)
	if err != nil {
		return nil, err
	}
	item, err := findWinner(wheel.Items, record.WinnerItemID, record.WinnerIdx)
	if err != nil {
		return nil, err
	}

	return &SpinResult{
		SpinID:    record.ID,
		WinnerIdx: record.WinnerIdx,
		Item:      item,
		Nonce:     record.Nonce,
		Sig:       sig,
		ExpiresAt: record.ExpiresAt,
	}, nil
}

func (s *wheelService) VerifySpin(spinID, wheelID string, winnerIdx int, itemsHash, nonce, sig string) bool {
	if len(s.hmacSecret) == 0 {
		return false
	}
	expected := signSpin(s.hmacSecret, spinID, wheelID, winnerIdx, itemsHash, nonce)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(sig)) == 1
}

func (s *wheelService) ListSpins(ctx context.Context, wheelID string, limit int) ([]*models.SpinRecord, error) {
	return s.repo.ListSpins(ctx, wheelID, limit)
}

func (s *wheelService) RecordEvent(ctx context.Context, params models.EventParams) (*models.EventRecord, error) {
	return s.repo.RecordEvent(ctx, params)
}

// findWinner resolves the decided item by ID, falling back to the recorded
// index (items can be edited between decision and read).
func findWinner(items []models.WheelItem, winnerItemID string, winnerIdx int) (*models.WheelItem, error) {
	for i := range items {
		if items[i].ID == winnerItemID {
			winner := items[i]
			return &winner, nil
		}
	}
	if winnerIdx < 0 || winnerIdx >= len(items) {
		return nil, errors.New("winner no longer on wheel")
	}
	winner := items[winnerIdx]
	return &winner, nil
}

// signSpin computes the opaque client sig per contract v1:
// HMAC_SHA256(HMAC_SECRET, spin_id|wheel_id|winner_idx|items_hash|nonce).
func signSpin(secret []byte, spinID, wheelID string, winnerIdx int, itemsHash, nonce string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(spinID + "|" + wheelID + "|" + strconv.Itoa(winnerIdx) + "|" + itemsHash + "|" + nonce))
	return hex.EncodeToString(mac.Sum(nil))
}

// ComputeItemsHash hashes the canonical item list per contract v1
// (service.proto:142-143):
//
//	sha256(join(sorted(item.id + "|" + option + "|" + weight + "|" + color), "\n"))
//
// Weight uses strconv 'g' format; the FE must serialize identically.
func ComputeItemsHash(items []models.WheelItem) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, item.ID+"|"+item.Option+"|"+
			strconv.FormatFloat(item.Weight, 'g', -1, 64)+"|"+item.Color)
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])
}
