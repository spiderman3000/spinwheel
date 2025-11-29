package repository

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"spinwheel/backend/pkg/models"
)

var (
	ErrNotFound = errors.New("not found")
)

type WheelRepository interface {
	CreateWheel(wheel *models.Wheel) (*models.Wheel, error)
	GetWheel(id string) (*models.Wheel, error)
	ListWheels() ([]*models.Wheel, error)
	UpdateWheel(wheel *models.Wheel) (*models.Wheel, error)
	DeleteWheel(id string) error
	SpinWheel(id string) (*models.WheelItem, error)
}

type InMemoryWheelRepository struct {
	wheels map[string]*models.Wheel
	mutex  sync.RWMutex
}

func NewInMemoryWheelRepository() *InMemoryWheelRepository {
	return &InMemoryWheelRepository{
		wheels: make(map[string]*models.Wheel),
	}
}

func (r *InMemoryWheelRepository) CreateWheel(wheel *models.Wheel) (*models.Wheel, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	wheel.ID = r.generateID()
	for i := range wheel.Items {
		wheel.Items[i].ID = r.generateID()
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
			wheel.Items[i].ID = r.generateID()
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

func (r *InMemoryWheelRepository) SpinWheel(id string) (*models.WheelItem, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	wheel, ok := r.wheels[id]
	if !ok {
		return nil, ErrNotFound
	}

	if len(wheel.Items) == 0 {
		return nil, errors.New("wheel has no items")
	}

	rand.Seed(time.Now().UnixNano())
	winningItem := wheel.Items[rand.Intn(len(wheel.Items))]
	return &winningItem, nil
}

func (r *InMemoryWheelRepository) generateID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
