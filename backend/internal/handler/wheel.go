package handler

import (
	"context"

	"spinwheel/backend/internal/service"
	"spinwheel/backend/pkg/models"
	spinwheelv1 "spinwheel/backend/gen/proto/spinwheel/v1"
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
		return nil, err
	}

	var createdItems []*spinwheelv1.WheelItem
	for _, item := range createdWheel.Items {
		createdItems = append(createdItems, &spinwheelv1.WheelItem{
			Id:     item.ID,
			Option: item.Option,
			Color:  item.Color,
		})
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
		return nil, err
	}

	var items []*spinwheelv1.WheelItem
	for _, item := range wheel.Items {
		items = append(items, &spinwheelv1.WheelItem{
			Id:     item.ID,
			Option: item.Option,
			Color:  item.Color,
		})
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
		return nil, err
	}

	var result []*spinwheelv1.Wheel
	for _, wheel := range wheels {
		var items []*spinwheelv1.WheelItem
		for _, item := range wheel.Items {
			items = append(items, &spinwheelv1.WheelItem{
				Id:     item.ID,
				Option: item.Option,
				Color:  item.Color,
			})
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
	var items []models.WheelItem
	for _, item := range req.Wheel.Items {
		items = append(items, models.WheelItem{
			ID:     item.Id,
			Option: item.Option,
			Color:  item.Color,
		})
	}

	wheel := &models.Wheel{
		ID:    req.Wheel.Id,
		Name:  req.Wheel.Name,
		Items: items,
	}

	updatedWheel, err := h.service.UpdateWheel(wheel)
	if err != nil {
		return nil, err
	}

	var updatedItems []*spinwheelv1.WheelItem
	for _, item := range updatedWheel.Items {
		updatedItems = append(updatedItems, &spinwheelv1.WheelItem{
			Id:     item.ID,
			Option: item.Option,
			Color:  item.Color,
		})
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
		return nil, err
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
		return nil, err
	}

	return &spinwheelv1.AddItemResponse{
		Item: &spinwheelv1.WheelItem{
			Id:     createdItem.ID,
			Option: createdItem.Option,
			Color:  createdItem.Color,
		},
	}, nil
}

func (h *WheelHandler) UpdateItem(ctx context.Context, req *spinwheelv1.UpdateItemRequest) (*spinwheelv1.UpdateItemResponse, error) {
	item := &models.WheelItem{
		ID:     req.Item.Id,
		Option: req.Item.Option,
		Color:  req.Item.Color,
	}

	updatedItem, err := h.service.UpdateItem(req.WheelId, item)
	if err != nil {
		return nil, err
	}

	return &spinwheelv1.UpdateItemResponse{
		Item: &spinwheelv1.WheelItem{
			Id:     updatedItem.ID,
			Option: updatedItem.Option,
			Color:  updatedItem.Color,
		},
	}, nil
}

func (h *WheelHandler) DeleteItem(ctx context.Context, req *spinwheelv1.DeleteItemRequest) (*spinwheelv1.DeleteItemResponse, error) {
	err := h.service.DeleteItem(req.WheelId, req.ItemId)
	if err != nil {
		return nil, err
	}
	return &spinwheelv1.DeleteItemResponse{}, nil
}

func (h *WheelHandler) SpinWheel(ctx context.Context, req *spinwheelv1.SpinWheelRequest) (*spinwheelv1.SpinWheelResponse, error) {
	wheel, err := h.service.GetWheel(req.WheelId)
	if err != nil {
		return nil, err
	}

	var items []*spinwheelv1.WheelItem
	for _, item := range wheel.Items {
		items = append(items, &spinwheelv1.WheelItem{
			Id:     item.ID,
			Option: item.Option,
			Color:  item.Color,
		})
	}

	return &spinwheelv1.SpinWheelResponse{
		Result: &spinwheelv1.SpinResult{
			SelectedItem: items[0], // Placeholder - actual logic would be in service
		},
	}, nil
}

func (h *WheelHandler) GetSpinHistory(ctx context.Context, req *spinwheelv1.GetSpinHistoryRequest) (*spinwheelv1.GetSpinHistoryResponse, error) {
	return &spinwheelv1.GetSpinHistoryResponse{
		Results: []*spinwheelv1.SpinResult{},
	}, nil
}
