package service

import (
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
	SpinWheel(id string) (*models.WheelItem, error)
}

type wheelService struct {
	repo repository.WheelRepository
}

func NewWheelService(repo repository.WheelRepository) WheelService {
	return &wheelService{
		repo: repo,
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

func (s *wheelService) SpinWheel(id string) (*models.WheelItem, error) {
	return s.repo.SpinWheel(id)
}

func (s *wheelService) AddItem(wheelID string, item *models.WheelItem) (*models.WheelItem, error) {
	// Placeholder - implement with repo method
	return item, nil
}

func (s *wheelService) UpdateItem(wheelID string, item *models.WheelItem) (*models.WheelItem, error) {
	// Placeholder - implement with repo method
	return item, nil
}

func (s *wheelService) DeleteItem(wheelID string, itemID string) error {
	// Placeholder - implement with repo method
	return nil
}
