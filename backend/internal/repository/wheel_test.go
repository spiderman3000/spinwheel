package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"spinwheel/backend/pkg/models"
)

func newTestWheel() *models.Wheel {
	return &models.Wheel{
		Name: "test wheel",
		Items: []models.WheelItem{
			{Option: "A", Color: "red", Weight: 3.0},
			{Option: "B", Color: "blue", Weight: 1.0},
		},
	}
}

func mustCreate(t *testing.T, repo WheelRepository) *models.Wheel {
	t.Helper()
	wheel, err := repo.CreateWheel(newTestWheel())
	if err != nil {
		t.Fatalf("CreateWheel: %v", err)
	}
	return wheel
}

func TestIDsAreUUID(t *testing.T) {
	repo := NewInMemoryWheelRepository()
	wheel := mustCreate(t, repo)

	if _, err := uuid.Parse(wheel.ID); err != nil {
		t.Fatalf("wheel ID %q is not a UUID: %v", wheel.ID, err)
	}
	for _, item := range wheel.Items {
		if _, err := uuid.Parse(item.ID); err != nil {
			t.Fatalf("item ID %q is not a UUID: %v", item.ID, err)
		}
	}

	spin, err := repo.RecordSpin(context.Background(), models.SpinParams{WheelID: wheel.ID})
	if err != nil {
		t.Fatalf("RecordSpin: %v", err)
	}
	if _, err := uuid.Parse(spin.ID); err != nil {
		t.Fatalf("spin ID %q is not a UUID: %v", spin.ID, err)
	}
	if len(spin.SeedServer) != 32 {
		t.Fatalf("seed_server len = %d, want 32", len(spin.SeedServer))
	}
	if len(spin.Nonce) != 32 {
		t.Fatalf("nonce len = %d, want 32 hex chars", len(spin.Nonce))
	}
	if spin.WinnerItemID != wheel.Items[spin.WinnerIdx].ID {
		t.Fatalf("WinnerItemID %q does not match items[%d].ID %q",
			spin.WinnerItemID, spin.WinnerIdx, wheel.Items[spin.WinnerIdx].ID)
	}
}

func TestWeightedPickRespectsWeights(t *testing.T) {
	repo := NewInMemoryWheelRepository()
	wheel := mustCreate(t, repo)

	const n = 2000
	aWins := 0
	for i := 0; i < n; i++ {
		spin, err := repo.RecordSpin(context.Background(), models.SpinParams{WheelID: wheel.ID})
		if err != nil {
			t.Fatalf("RecordSpin: %v", err)
		}
		if spin.WinnerIdx < 0 || spin.WinnerIdx > 1 {
			t.Fatalf("WinnerIdx = %d, want 0 or 1", spin.WinnerIdx)
		}
		if wheel.Items[spin.WinnerIdx].Option == "A" {
			aWins++
		}
	}
	share := float64(aWins) / n // expect ~0.75 for weights 3:1
	if share < 0.60 || share > 0.90 {
		t.Fatalf("A share = %.3f over %d spins, want in [0.60, 0.90] for 3:1 weights", share, n)
	}
}

func TestItemCRUD(t *testing.T) {
	repo := NewInMemoryWheelRepository()
	wheel := mustCreate(t, repo)

	added, err := repo.AddItem(wheel.ID, &models.WheelItem{Option: "C", Color: "green", Weight: 2.0})
	if err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	if _, err := uuid.Parse(added.ID); err != nil {
		t.Fatalf("added item ID %q is not a UUID: %v", added.ID, err)
	}

	added.Option = "C2"
	updated, err := repo.UpdateItem(wheel.ID, added)
	if err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}
	if updated.Option != "C2" {
		t.Fatalf("Option = %q, want C2", updated.Option)
	}

	if err := repo.DeleteItem(wheel.ID, added.ID); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
	if err := repo.DeleteItem(wheel.ID, added.ID); err != ErrNotFound {
		t.Fatalf("second DeleteItem = %v, want ErrNotFound", err)
	}
	if _, err := repo.AddItem("no-such-wheel", &models.WheelItem{}); err != ErrNotFound {
		t.Fatalf("AddItem on missing wheel = %v, want ErrNotFound", err)
	}
	if _, err := repo.UpdateItem(wheel.ID, &models.WheelItem{ID: "no-such-item"}); err != ErrNotFound {
		t.Fatalf("UpdateItem on missing item = %v, want ErrNotFound", err)
	}
	if err := repo.DeleteItem("no-such-wheel", added.ID); err != ErrNotFound {
		t.Fatalf("DeleteItem on missing wheel = %v, want ErrNotFound", err)
	}
}

func TestSpinWheelErrors(t *testing.T) {
	repo := NewInMemoryWheelRepository()

	if _, err := repo.SpinWheel("no-such-wheel"); err != ErrNotFound {
		t.Fatalf("SpinWheel on missing wheel = %v, want ErrNotFound", err)
	}
	empty, err := repo.CreateWheel(&models.Wheel{Name: "empty"})
	if err != nil {
		t.Fatalf("CreateWheel: %v", err)
	}
	if _, err := repo.SpinWheel(empty.ID); err == nil {
		t.Fatal("SpinWheel on empty wheel: want error, got nil")
	}
	if _, err := repo.RecordSpin(context.Background(), models.SpinParams{WheelID: empty.ID}); err == nil {
		t.Fatal("RecordSpin on empty wheel: want error, got nil")
	}
}

func TestRecordEventIdempotency(t *testing.T) {
	repo := NewInMemoryWheelRepository()
	ctx := context.Background()
	ts := time.Now().UTC().Truncate(time.Millisecond)

	params := models.EventParams{
		SessionID: "sess-1", Type: "PAGEVIEW", Path: "/w/1", ClientTS: ts,
	}
	first, err := repo.RecordEvent(ctx, params)
	if err != nil {
		t.Fatalf("RecordEvent: %v", err)
	}
	second, err := repo.RecordEvent(ctx, params)
	if err != nil {
		t.Fatalf("RecordEvent redelivery: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("redelivery returned new ID %q, want %q", second.ID, first.ID)
	}

	params.Type = "SPIN_START"
	third, err := repo.RecordEvent(ctx, params)
	if err != nil {
		t.Fatalf("RecordEvent other type: %v", err)
	}
	if third.ID == first.ID {
		t.Fatal("different event type reused the same row")
	}
}

func TestListSpins(t *testing.T) {
	repo := NewInMemoryWheelRepository()
	ctx := context.Background()
	wheel := mustCreate(t, repo)

	if _, err := repo.ListSpins(ctx, "no-such-wheel", 10); err != ErrNotFound {
		t.Fatalf("ListSpins on missing wheel = %v, want ErrNotFound", err)
	}

	for i := 0; i < 3; i++ {
		if _, err := repo.RecordSpin(ctx, models.SpinParams{WheelID: wheel.ID}); err != nil {
			t.Fatalf("RecordSpin: %v", err)
		}
	}

	spins, err := repo.ListSpins(ctx, wheel.ID, 0)
	if err != nil {
		t.Fatalf("ListSpins: %v", err)
	}
	if len(spins) != 3 {
		t.Fatalf("len = %d, want 3", len(spins))
	}
	for i := 1; i < len(spins); i++ {
		if spins[i].CreatedAt.After(spins[i-1].CreatedAt) {
			t.Fatal("spins are not newest-first")
		}
	}

	spins, err = repo.ListSpins(ctx, wheel.ID, 2)
	if err != nil {
		t.Fatalf("ListSpins limit: %v", err)
	}
	if len(spins) != 2 {
		t.Fatalf("len = %d, want 2", len(spins))
	}
}
