package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"spinwheel/backend/pkg/models"
)

// TestPostgresRepository exercises the full Postgres flow against a real
// database. It only runs when TEST_DATABASE_URL is set (Neon direct URL),
// so unit CI stays hermetic:
//
//	TEST_DATABASE_URL="$DIRECT_URL" go test ./internal/repository/ -run TestPostgres -v
func TestPostgresRepository(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	repo, err := NewPostgresWheelRepository(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	// Close via Cleanup (not defer): on t.Fatalf, deferred calls run before
	// Cleanup functions, which would close the pool out from under cleanup.
	t.Cleanup(repo.Close)

	// Wheel CRUD ------------------------------------------------------
	wheel, err := repo.CreateWheel(&models.Wheel{
		Name: "pg-test-" + time.Now().UTC().Format("150405"),
		Items: []models.WheelItem{
			{Option: "A", Color: "red", Weight: 3.0},
			{Option: "B", Color: "blue", Weight: 1.0},
		},
	})
	if err != nil {
		t.Fatalf("CreateWheel: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.DeleteWheel(wheel.ID); err != nil {
			t.Errorf("cleanup DeleteWheel: %v", err)
		}
	})
	if _, err := uuid.Parse(wheel.ID); err != nil {
		t.Fatalf("wheel ID %q is not a UUID: %v", wheel.ID, err)
	}

	got, err := repo.GetWheel(wheel.ID)
	if err != nil {
		t.Fatalf("GetWheel: %v", err)
	}
	if len(got.Items) != 2 || got.Items[0].Option != "A" || got.Items[1].Weight != 1.0 {
		t.Fatalf("GetWheel items = %+v", got.Items)
	}
	if _, err := repo.GetWheel(uuid.NewString()); err != ErrNotFound {
		t.Fatalf("GetWheel missing = %v, want ErrNotFound", err)
	}

	// Item CRUD --------------------------------------------------------
	added, err := repo.AddItem(wheel.ID, &models.WheelItem{Option: "C", Color: "green", Weight: 2.0})
	if err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	added.Option = "C2"
	if _, err := repo.UpdateItem(wheel.ID, added); err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}
	if err := repo.DeleteItem(wheel.ID, added.ID); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
	if err := repo.DeleteItem(wheel.ID, added.ID); err != ErrNotFound {
		t.Fatalf("second DeleteItem = %v, want ErrNotFound", err)
	}

	// Spin: read-only pick + persisted decisions ------------------------
	pick, err := repo.SpinWheel(wheel.ID)
	if err != nil {
		t.Fatalf("SpinWheel: %v", err)
	}
	if pick.Option != "A" && pick.Option != "B" {
		t.Fatalf("SpinWheel option = %q", pick.Option)
	}

	const n = 200
	aWins := 0
	var lastSpinID string
	for i := 0; i < n; i++ {
		spin, err := repo.RecordSpin(ctx, models.SpinParams{
			WheelID: wheel.ID, SessionID: "sess-pg",
		})
		if err != nil {
			t.Fatalf("RecordSpin: %v", err)
		}
		if spin.WinnerIdx < 0 || spin.WinnerIdx > 1 {
			t.Fatalf("WinnerIdx = %d", spin.WinnerIdx)
		}
		if len(spin.SeedServer) != 32 || len(spin.Nonce) != 32 {
			t.Fatalf("seed len = %d, nonce len = %d", len(spin.SeedServer), len(spin.Nonce))
		}
		if spin.WinnerItemID != got.Items[spin.WinnerIdx].ID {
			t.Fatalf("WinnerItemID %q != items[%d].ID %q",
				spin.WinnerItemID, spin.WinnerIdx, got.Items[spin.WinnerIdx].ID)
		}
		if got.Items[spin.WinnerIdx].Option == "A" {
			aWins++
		}
		lastSpinID = spin.ID
	}
	share := float64(aWins) / n // expect ~0.75 for 3:1
	if share < 0.60 || share > 0.90 {
		t.Fatalf("A share = %.3f over %d spins, want [0.60, 0.90]", share, n)
	}

	history, err := repo.ListSpins(ctx, wheel.ID, 10)
	if err != nil {
		t.Fatalf("ListSpins: %v", err)
	}
	if len(history) != 10 {
		t.Fatalf("len = %d, want 10", len(history))
	}
	for i := 1; i < len(history); i++ {
		if history[i].CreatedAt.After(history[i-1].CreatedAt) {
			t.Fatal("spins are not newest-first")
		}
	}
	if history[0].Sig != "" {
		t.Fatalf("Sig = %q, want empty (service fills it in step 2)", history[0].Sig)
	}

	// Events: idempotent insert ----------------------------------------
	ts := time.Now().UTC().Truncate(time.Microsecond)
	params := models.EventParams{
		SessionID: "sess-pg", Type: "SPIN_END", WheelID: wheel.ID,
		SpinID: lastSpinID, Sig: "sig-test", ClientTS: ts,
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
		t.Fatalf("redelivery ID %q != %q", second.ID, first.ID)
	}
	if second.WheelID != wheel.ID || second.SpinID != lastSpinID {
		t.Fatalf("event keys not round-tripped: %+v", second)
	}
}
