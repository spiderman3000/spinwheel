package service

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"spinwheel/backend/internal/repository"
	"spinwheel/backend/pkg/models"
)

const testSecret = "test-hmac-secret"

func newTestService() (WheelService, *models.Wheel) {
	repo := repository.NewInMemoryWheelRepository()
	wheel, err := repo.CreateWheel(&models.Wheel{
		Name: "svc-test",
		Items: []models.WheelItem{
			{Option: "A", Color: "red", Weight: 1.0},
			{Option: "B", Color: "blue", Weight: 1.0},
		},
	})
	if err != nil {
		panic(err)
	}
	return NewWheelService(repo, testSecret), wheel
}

func TestSpinWheelSecureRoundTrip(t *testing.T) {
	svc, wheel := newTestService()
	ctx := context.Background()

	result, err := svc.SpinWheelSecure(ctx, wheel.ID, SpinOptions{
		ClientSeed: "client-seed-1",
		ItemsHash:  "hash-from-client",
		SessionID:  "sess-1",
	})
	if err != nil {
		t.Fatalf("SpinWheelSecure: %v", err)
	}
	if _, err := uuid.Parse(result.SpinID); err != nil {
		t.Fatalf("SpinID %q is not a UUID: %v", result.SpinID, err)
	}
	if result.Sig == "" || result.Nonce == "" {
		t.Fatal("Sig and Nonce must be set")
	}
	if result.Item.ID != wheel.Items[result.WinnerIdx].ID {
		t.Fatalf("Item.ID %q != items[%d].ID %q",
			result.Item.ID, result.WinnerIdx, wheel.Items[result.WinnerIdx].ID)
	}

	if !svc.VerifySpin(result.SpinID, wheel.ID, result.WinnerIdx, "hash-from-client", result.Nonce, result.Sig) {
		t.Fatal("VerifySpin rejected a genuine decision")
	}

	// Sig must be persisted on the audit row.
	history, err := svc.ListSpins(ctx, wheel.ID, 10)
	if err != nil {
		t.Fatalf("ListSpins: %v", err)
	}
	if len(history) != 1 || history[0].Sig != result.Sig {
		t.Fatalf("stored sig not persisted: %+v", history)
	}
}

func TestVerifySpinRejectsTampering(t *testing.T) {
	svc, wheel := newTestService()
	ctx := context.Background()

	result, err := svc.SpinWheelSecure(ctx, wheel.ID, SpinOptions{ItemsHash: "h"})
	if err != nil {
		t.Fatalf("SpinWheelSecure: %v", err)
	}
	ok := func(winnerIdx int, itemsHash, nonce, sig string) bool {
		return svc.VerifySpin(result.SpinID, wheel.ID, winnerIdx, itemsHash, nonce, sig)
	}

	if ok(result.WinnerIdx+1, "h", result.Nonce, result.Sig) {
		t.Error("flipped winner_index verified")
	}
	if ok(result.WinnerIdx, "other-hash", result.Nonce, result.Sig) {
		t.Error("swapped items_hash verified")
	}
	if ok(result.WinnerIdx, "h", "00"+result.Nonce[2:], result.Sig) {
		t.Error("swapped nonce verified")
	}
	if ok(result.WinnerIdx, "h", result.Nonce, "00"+result.Sig[2:]) {
		t.Error("forged sig verified")
	}
	if ok(result.WinnerIdx, "h", result.Nonce, strings.ToUpper(result.Sig)) {
		t.Error("case-mutated sig verified")
	}
	// Wrong secret must fail.
	other := NewWheelService(repository.NewInMemoryWheelRepository(), "other-secret")
	if other.VerifySpin(result.SpinID, wheel.ID, result.WinnerIdx, "h", result.Nonce, result.Sig) {
		t.Error("sig verified under a different secret")
	}
}

func TestSpinWheelSecureRequiresSecret(t *testing.T) {
	repo := repository.NewInMemoryWheelRepository()
	wheel, err := repo.CreateWheel(&models.Wheel{
		Name:  "x",
		Items: []models.WheelItem{{Option: "A"}},
	})
	if err != nil {
		t.Fatalf("CreateWheel: %v", err)
	}
	svc := NewWheelService(repo, "")
	if _, err := svc.SpinWheelSecure(context.Background(), wheel.ID, SpinOptions{}); err == nil {
		t.Fatal("SpinWheelSecure without secret: want error, got nil")
	}
	if svc.VerifySpin("s", wheel.ID, 0, "h", "n", "sig") {
		t.Fatal("VerifySpin without secret: want false")
	}
}

func TestItemDelegation(t *testing.T) {
	svc, wheel := newTestService()

	added, err := svc.AddItem(wheel.ID, &models.WheelItem{Option: "C", Weight: 2.0})
	if err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	if _, err := uuid.Parse(added.ID); err != nil {
		t.Fatalf("AddItem ID %q is not a UUID: %v", added.ID, err)
	}

	added.Option = "C2"
	if _, err := svc.UpdateItem(wheel.ID, added); err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}
	if err := svc.DeleteItem(wheel.ID, added.ID); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
	if _, err := svc.AddItem("no-such-wheel", &models.WheelItem{}); err != repository.ErrNotFound {
		t.Fatalf("AddItem missing wheel = %v, want ErrNotFound", err)
	}
}

func TestComputeItemsHash(t *testing.T) {
	a := []models.WheelItem{
		{ID: "1", Option: "A", Color: "red", Weight: 3.0},
		{ID: "2", Option: "B", Color: "blue", Weight: 1.0},
	}
	b := []models.WheelItem{
		{ID: "2", Option: "B", Color: "blue", Weight: 1.0},
		{ID: "1", Option: "A", Color: "red", Weight: 3.0},
	}
	if ComputeItemsHash(a) != ComputeItemsHash(b) {
		t.Fatal("hash must be order-independent (items are sorted)")
	}
	c := []models.WheelItem{
		{ID: "1", Option: "A", Color: "red", Weight: 2.0},
		{ID: "2", Option: "B", Color: "blue", Weight: 1.0},
	}
	if ComputeItemsHash(a) == ComputeItemsHash(c) {
		t.Fatal("hash must change when a weight changes")
	}
	if len(ComputeItemsHash(a)) != 64 {
		t.Fatal("hash must be 64 hex chars (sha256)")
	}
}

func TestFindWinnerFallback(t *testing.T) {
	items := []models.WheelItem{
		{ID: "a", Option: "A"},
		{ID: "b", Option: "B"},
	}
	// Exact ID match wins even if the index drifted.
	got, err := findWinner(items, "b", 0)
	if err != nil || got.ID != "b" {
		t.Fatalf("findWinner by ID = %+v, %v", got, err)
	}
	// Deleted item falls back to the recorded index.
	got, err = findWinner(items, "gone", 1)
	if err != nil || got.ID != "b" {
		t.Fatalf("findWinner fallback = %+v, %v", got, err)
	}
	if _, err := findWinner(items, "gone", 7); err == nil {
		t.Fatal("out-of-range fallback: want error")
	}
}

func TestRecordEventPassthrough(t *testing.T) {
	svc, _ := newTestService()
	rec, err := svc.RecordEvent(context.Background(), models.EventParams{
		SessionID: "s", Type: "PAGEVIEW", Path: "/",
	})
	if err != nil {
		t.Fatalf("RecordEvent: %v", err)
	}
	if rec.ID == "" {
		t.Fatal("RecordEvent returned empty ID")
	}
}
