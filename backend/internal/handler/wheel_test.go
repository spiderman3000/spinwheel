package handler

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	spinwheelv1 "spinwheel/backend/gen/proto/spinwheel/v1"
	"spinwheel/backend/internal/repository"
	"spinwheel/backend/internal/service"
	"spinwheel/backend/pkg/models"
)

const testSecret = "test-hmac-secret"

func newTestHandler() (service.WheelService, *WheelHandler, *models.Wheel) {
	repo := repository.NewInMemoryWheelRepository()
	svc := service.NewWheelService(repo, testSecret)
	wheel, err := svc.CreateWheel(&models.Wheel{
		Name: "hdl-test",
		Items: []models.WheelItem{
			{Option: "A", Color: "red", Weight: 3.0},
			{Option: "B", Color: "blue", Weight: 1.0},
		},
	})
	if err != nil {
		panic(err)
	}
	return svc, NewWheelHandler(svc), wheel
}

func codeOf(t *testing.T, err error) codes.Code {
	t.Helper()
	if err == nil {
		t.Fatal("want error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error %v is not a gRPC status", err)
	}
	return st.Code()
}

func TestSpinWheelSecure(t *testing.T) {
	svc, hdl, wheel := newTestHandler()
	ctx := context.Background()

	hash := service.ComputeItemsHash(wheel.Items)
	resp, err := hdl.SpinWheel(ctx, &spinwheelv1.SpinWheelRequest{
		WheelId:    wheel.ID,
		ClientSeed: "client-seed-1",
		ItemsHash:  hash,
	})
	if err != nil {
		t.Fatalf("SpinWheel: %v", err)
	}
	if resp.SpinId == "" || resp.Sig == "" || resp.Nonce == "" {
		t.Fatal("SpinId/Sig/Nonce must be set")
	}
	if resp.WinnerItem == nil || resp.WinnerItem.Weight != wheel.Items[resp.WinnerIndex].Weight {
		t.Fatalf("WinnerItem weight not mapped: %+v", resp.WinnerItem)
	}
	if resp.ExpiresAt == nil {
		t.Fatal("ExpiresAt must be set")
	}
	if !svc.VerifySpin(resp.SpinId, wheel.ID, int(resp.WinnerIndex), hash, resp.Nonce, resp.Sig) {
		t.Fatal("response sig does not verify")
	}
	// Deprecated compat wrapper is filled too.
	if resp.Result == nil || resp.Result.Id != resp.SpinId {
		t.Fatal("deprecated Result wrapper not filled")
	}
}

func TestSpinWheelRejectsStaleOrEmptyHash(t *testing.T) {
	_, hdl, wheel := newTestHandler()
	ctx := context.Background()

	_, err := hdl.SpinWheel(ctx, &spinwheelv1.SpinWheelRequest{WheelId: wheel.ID})
	if codeOf(t, err) != codes.FailedPrecondition {
		t.Fatalf("empty hash code = %v, want FailedPrecondition", err)
	}
	_, err = hdl.SpinWheel(ctx, &spinwheelv1.SpinWheelRequest{WheelId: wheel.ID, ItemsHash: "stale"})
	if codeOf(t, err) != codes.FailedPrecondition {
		t.Fatalf("stale hash code = %v, want FailedPrecondition", err)
	}
	_, err = hdl.SpinWheel(ctx, &spinwheelv1.SpinWheelRequest{WheelId: "no-such-wheel", ItemsHash: "x"})
	if codeOf(t, err) != codes.NotFound {
		t.Fatalf("missing wheel code = %v, want NotFound", err)
	}
	_, err = hdl.SpinWheel(ctx, &spinwheelv1.SpinWheelRequest{})
	if codeOf(t, err) != codes.InvalidArgument {
		t.Fatalf("empty wheel_id code = %v, want InvalidArgument", err)
	}
}

func TestGetSpinHistoryPagination(t *testing.T) {
	svc, hdl, wheel := newTestHandler()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if _, err := svc.SpinWheelSecure(ctx, wheel.ID, service.SpinOptions{ItemsHash: "h"}); err != nil {
			t.Fatalf("SpinWheelSecure: %v", err)
		}
	}

	page1, err := hdl.GetSpinHistory(ctx, &spinwheelv1.GetSpinHistoryRequest{WheelId: wheel.ID, PageSize: 2})
	if err != nil {
		t.Fatalf("GetSpinHistory: %v", err)
	}
	if len(page1.Results) != 2 || page1.NextPageToken != "2" {
		t.Fatalf("page1 = %d results token %q, want 2 + \"2\"", len(page1.Results), page1.NextPageToken)
	}
	for _, r := range page1.Results {
		if r.SelectedItem == nil || r.SpunAt == nil {
			t.Fatalf("history row missing item/time: %+v", r)
		}
	}

	page2, err := hdl.GetSpinHistory(ctx, &spinwheelv1.GetSpinHistoryRequest{WheelId: wheel.ID, PageSize: 2, PageToken: page1.NextPageToken})
	if err != nil {
		t.Fatalf("GetSpinHistory page 2: %v", err)
	}
	if len(page2.Results) != 1 || page2.NextPageToken != "" {
		t.Fatalf("page2 = %d results token %q, want 1 + empty", len(page2.Results), page2.NextPageToken)
	}

	if _, err := hdl.GetSpinHistory(ctx, &spinwheelv1.GetSpinHistoryRequest{WheelId: wheel.ID, PageToken: "bogus"}); codeOf(t, err) != codes.InvalidArgument {
		t.Fatalf("bad token code = %v, want InvalidArgument", err)
	}
	if _, err := hdl.GetSpinHistory(ctx, &spinwheelv1.GetSpinHistoryRequest{WheelId: "no-such-wheel"}); codeOf(t, err) != codes.NotFound {
		t.Fatalf("missing wheel code = %v, want NotFound", err)
	}
}

func TestRecordEvent(t *testing.T) {
	_, hdl, wheel := newTestHandler()
	ctx := context.Background()

	_, err := hdl.RecordEvent(ctx, &spinwheelv1.RecordEventRequest{
		Type:      spinwheelv1.EventType_PAGEVIEW,
		Path:      "/play/abc",
		SessionId: "sess-1",
	})
	if err != nil {
		t.Fatalf("RecordEvent PAGEVIEW: %v", err)
	}
	// Redelivery is a no-op, not an error.
	_, err = hdl.RecordEvent(ctx, &spinwheelv1.RecordEventRequest{
		Type:      spinwheelv1.EventType_PAGEVIEW,
		Path:      "/play/abc",
		SessionId: "sess-1",
	})
	if err != nil {
		t.Fatalf("RecordEvent redelivery: %v", err)
	}

	if _, err := hdl.RecordEvent(ctx, &spinwheelv1.RecordEventRequest{}); codeOf(t, err) != codes.InvalidArgument {
		t.Fatalf("unspecified type code = %v, want InvalidArgument", err)
	}
	_, err = hdl.RecordEvent(ctx, &spinwheelv1.RecordEventRequest{
		Type: spinwheelv1.EventType_SPIN_END, WheelId: wheel.ID,
		SpinId: "not-a-uuid", SessionId: "sess-1",
	})
	if codeOf(t, err) != codes.InvalidArgument {
		t.Fatalf("bad spin_id code = %v, want InvalidArgument", err)
	}
}
