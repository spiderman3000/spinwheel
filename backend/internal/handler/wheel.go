package handler

import (
	"context"
	"errors"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"

	spinwheelv1 "spinwheel/backend/gen/proto/spinwheel/v1"
	"spinwheel/backend/internal/middleware"
	"spinwheel/backend/internal/repository"
	"spinwheel/backend/internal/service"
	"spinwheel/backend/pkg/models"
)

type WheelHandler struct {
	spinwheelv1.UnimplementedWheelServiceServer
	service service.WheelService
}

func NewWheelHandler(service service.WheelService) *WheelHandler {
	return &WheelHandler{
		service: service,
	}
}

func (h *WheelHandler) CreateWheel(ctx context.Context, req *spinwheelv1.CreateWheelRequest) (*spinwheelv1.CreateWheelResponse, error) {
	var items []models.WheelItem
	for _, itemName := range req.InitialItems {
		items = append(items, models.WheelItem{
			Option: itemName,
		})
	}

	wheel := &models.Wheel{
		Name:  req.Name,
		Items: items,
	}

	createdWheel, err := h.service.CreateWheel(wheel)
	if err != nil {
		return nil, toStatus(err)
	}

	var createdItems []*spinwheelv1.WheelItem
	for _, item := range createdWheel.Items {
		createdItems = append(createdItems, toProtoItem(item))
	}

	return &spinwheelv1.CreateWheelResponse{
		Wheel: &spinwheelv1.Wheel{
			Id:    createdWheel.ID,
			Name:  createdWheel.Name,
			Items: createdItems,
		},
	}, nil
}

func (h *WheelHandler) GetWheel(ctx context.Context, req *spinwheelv1.GetWheelRequest) (*spinwheelv1.GetWheelResponse, error) {
	wheel, err := h.service.GetWheel(req.Id)
	if err != nil {
		return nil, toStatus(err)
	}

	var items []*spinwheelv1.WheelItem
	for _, item := range wheel.Items {
		items = append(items, toProtoItem(item))
	}

	return &spinwheelv1.GetWheelResponse{
		Wheel: &spinwheelv1.Wheel{
			Id:    wheel.ID,
			Name:  wheel.Name,
			Items: items,
		},
	}, nil
}

func (h *WheelHandler) ListWheels(ctx context.Context, req *spinwheelv1.ListWheelsRequest) (*spinwheelv1.ListWheelsResponse, error) {
	wheels, err := h.service.ListWheels()
	if err != nil {
		return nil, toStatus(err)
	}

	var result []*spinwheelv1.Wheel
	for _, wheel := range wheels {
		var items []*spinwheelv1.WheelItem
		for _, item := range wheel.Items {
			items = append(items, toProtoItem(item))
		}
		result = append(result, &spinwheelv1.Wheel{
			Id:    wheel.ID,
			Name:  wheel.Name,
			Items: items,
		})
	}

	return &spinwheelv1.ListWheelsResponse{
		Wheels: result,
	}, nil
}

func (h *WheelHandler) UpdateWheel(ctx context.Context, req *spinwheelv1.UpdateWheelRequest) (*spinwheelv1.UpdateWheelResponse, error) {
	if req.Wheel == nil {
		return nil, status.Error(codes.InvalidArgument, "wheel is required")
	}
	var items []models.WheelItem
	for _, item := range req.Wheel.Items {
		items = append(items, models.WheelItem{
			ID:     item.Id,
			Option: item.Option,
			Color:  item.Color,
			Weight: item.Weight,
		})
	}

	wheel := &models.Wheel{
		ID:    req.Wheel.Id,
		Name:  req.Wheel.Name,
		Items: items,
	}

	updatedWheel, err := h.service.UpdateWheel(wheel)
	if err != nil {
		return nil, toStatus(err)
	}

	var updatedItems []*spinwheelv1.WheelItem
	for _, item := range updatedWheel.Items {
		updatedItems = append(updatedItems, toProtoItem(item))
	}

	return &spinwheelv1.UpdateWheelResponse{
		Wheel: &spinwheelv1.Wheel{
			Id:    updatedWheel.ID,
			Name:  updatedWheel.Name,
			Items: updatedItems,
		},
	}, nil
}

func (h *WheelHandler) DeleteWheel(ctx context.Context, req *spinwheelv1.DeleteWheelRequest) (*spinwheelv1.DeleteWheelResponse, error) {
	err := h.service.DeleteWheel(req.Id)
	if err != nil {
		return nil, toStatus(err)
	}
	return &spinwheelv1.DeleteWheelResponse{}, nil
}

func (h *WheelHandler) AddItem(ctx context.Context, req *spinwheelv1.AddItemRequest) (*spinwheelv1.AddItemResponse, error) {
	item := &models.WheelItem{
		Option: req.Option,
		Color:  req.Color,
		Weight: req.Weight,
	}

	createdItem, err := h.service.AddItem(req.WheelId, item)
	if err != nil {
		return nil, toStatus(err)
	}

	return &spinwheelv1.AddItemResponse{
		Item: toProtoItem(*createdItem),
	}, nil
}

func (h *WheelHandler) UpdateItem(ctx context.Context, req *spinwheelv1.UpdateItemRequest) (*spinwheelv1.UpdateItemResponse, error) {
	if req.Item == nil {
		return nil, status.Error(codes.InvalidArgument, "item is required")
	}
	item := &models.WheelItem{
		ID:     req.Item.Id,
		Option: req.Item.Option,
		Color:  req.Item.Color,
		Weight: req.Item.Weight,
	}

	updatedItem, err := h.service.UpdateItem(req.WheelId, item)
	if err != nil {
		return nil, toStatus(err)
	}

	return &spinwheelv1.UpdateItemResponse{
		Item: toProtoItem(*updatedItem),
	}, nil
}

func (h *WheelHandler) DeleteItem(ctx context.Context, req *spinwheelv1.DeleteItemRequest) (*spinwheelv1.DeleteItemResponse, error) {
	err := h.service.DeleteItem(req.WheelId, req.ItemId)
	if err != nil {
		return nil, toStatus(err)
	}
	return &spinwheelv1.DeleteItemResponse{}, nil
}

func (h *WheelHandler) SpinWheel(ctx context.Context, req *spinwheelv1.SpinWheelRequest) (*spinwheelv1.SpinWheelResponse, error) {
	if req.WheelId == "" {
		return nil, status.Error(codes.InvalidArgument, "wheel_id is required")
	}

	wheel, err := h.service.GetWheel(req.WheelId)
	if err != nil {
		return nil, toStatus(err)
	}

	// Strict freshness check (contract v1): empty or stale items_hash is
	// rejected — a spin drawn against an unknown list could never verify.
	if req.ItemsHash == "" {
		return nil, status.Error(codes.FailedPrecondition, "items_hash is required (wheel_changed)")
	}
	if want := service.ComputeItemsHash(wheel.Items); want != req.ItemsHash {
		return nil, status.Error(codes.FailedPrecondition, "items_hash mismatch (wheel_changed)")
	}

	result, err := h.service.SpinWheelSecure(ctx, req.WheelId, service.SpinOptions{
		ClientSeed: req.ClientSeed,
		ItemsHash:  req.ItemsHash,
		SessionID:  sessionFromContext(ctx),
		// IPHash/UA/CFRay stay empty on gRPC; the HTTP gateway (step 4)
		// fills them from the connection and headers.
	})
	if err != nil {
		return nil, toStatus(err)
	}

	winnerItem := toProtoItem(*result.Item)
	return &spinwheelv1.SpinWheelResponse{
		Result: &spinwheelv1.SpinResult{
			Id:           result.SpinID,
			WheelId:      req.WheelId,
			SelectedItem: winnerItem,
		},
		SpinId:      result.SpinID,
		WinnerIndex: int32(result.WinnerIdx),
		WinnerItem:  winnerItem,
		Nonce:       result.Nonce,
		Sig:         result.Sig,
		ExpiresAt:   timestamppb.New(result.ExpiresAt),
	}, nil
}

func (h *WheelHandler) GetSpinHistory(ctx context.Context, req *spinwheelv1.GetSpinHistoryRequest) (*spinwheelv1.GetSpinHistoryResponse, error) {
	offset := 0
	if req.PageToken != "" {
		var err error
		offset, err = strconv.Atoi(req.PageToken)
		if err != nil || offset < 0 {
			return nil, status.Error(codes.InvalidArgument, "invalid page_token")
		}
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	// Fetch one extra row to know whether another page exists.
	records, err := h.service.ListSpins(ctx, req.WheelId, pageSize+1)
	if err != nil {
		return nil, toStatus(err)
	}
	// ListSpins has no offset, so page locally (histories are small;
	// server-side cursors are a later optimization).
	if offset > len(records) {
		offset = len(records)
	}
	end := offset + pageSize
	hasMore := false
	if end < len(records) {
		hasMore = true
	} else {
		end = len(records)
	}
	page := records[offset:end]

	results := make([]*spinwheelv1.SpinResult, 0, len(page))
	for _, record := range page {
		item, err := h.service.GetWheel(record.WheelID)
		if err != nil {
			return nil, toStatus(err)
		}
		var selected *spinwheelv1.WheelItem
		for _, it := range item.Items {
			if it.ID == record.WinnerItemID {
				selected = toProtoItem(it)
				break
			}
		}
		if selected == nil {
			// Item edited away since the spin; index fallback.
			if record.WinnerIdx >= 0 && record.WinnerIdx < len(item.Items) {
				selected = toProtoItem(item.Items[record.WinnerIdx])
			}
		}
		results = append(results, &spinwheelv1.SpinResult{
			Id:           record.ID,
			WheelId:      record.WheelID,
			SelectedItem: selected,
			SpunAt:       timestamppb.New(record.CreatedAt),
		})
	}

	resp := &spinwheelv1.GetSpinHistoryResponse{
		Results: results,
		// Approximate: len of this page. Accurate totals need a
		// CountSpins query (backlog).
		TotalCount: int32(len(results)),
	}
	if hasMore {
		resp.NextPageToken = strconv.Itoa(offset + pageSize)
	}
	return resp, nil
}

func (h *WheelHandler) RecordEvent(ctx context.Context, req *spinwheelv1.RecordEventRequest) (*spinwheelv1.RecordEventResponse, error) {
	var eventType string
	switch req.Type {
	case spinwheelv1.EventType_PAGEVIEW:
		eventType = "PAGEVIEW"
	case spinwheelv1.EventType_SPIN_START:
		eventType = "SPIN_START"
	case spinwheelv1.EventType_SPIN_END:
		eventType = "SPIN_END"
	default:
		return nil, status.Error(codes.InvalidArgument, "event type is required")
	}

	if req.WheelId != "" {
		if _, err := uuid.Parse(req.WheelId); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid wheel_id")
		}
	}
	if req.SpinId != "" {
		if _, err := uuid.Parse(req.SpinId); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid spin_id")
		}
	}

	params := models.EventParams{
		SessionID: sessionFromContext(ctx, req.SessionId),
		Type:      eventType,
		WheelID:   req.WheelId,
		SpinID:    req.SpinId,
		Path:      req.Path,
		Sig:       req.Sig,
	}
	if req.ClientTs != nil {
		params.ClientTS = req.ClientTs.AsTime()
	}

	if _, err := h.service.RecordEvent(ctx, params); err != nil {
		return nil, toStatus(err)
	}
	return &spinwheelv1.RecordEventResponse{}, nil
}

// toProtoItem maps the full item, including Weight (previously dropped,
// which silently turned weighted spins uniform on FE-driven renders).
func toProtoItem(item models.WheelItem) *spinwheelv1.WheelItem {
	return &spinwheelv1.WheelItem{
		Id:     item.ID,
		Option: item.Option,
		Color:  item.Color,
		Weight: item.Weight,
	}
}

// sessionFromContext resolves the anon session: explicit value first,
// then the x-user-id gRPC metadata fallback (contract v1, grpcurl/tests).
func sessionFromContext(ctx context.Context, explicit ...string) string {
	for _, s := range explicit {
		if s != "" {
			return s
		}
	}
	if userID, ok := middleware.UserIDFromContext(ctx); ok {
		return userID
	}
	return ""
}

func toStatus(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}
