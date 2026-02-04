package service

import (
	"context"
	"fiberbackend/internal/model"
	"fiberbackend/internal/repository"
	"time"
)

type HoldingService interface {
	Create(ctx context.Context, userID string, dto *model.CreateHoldingDTO) (*model.HoldingResponse, error)
	GetByID(ctx context.Context, id int64, userID string) (*model.HoldingResponse, error)
	ListByUserID(ctx context.Context, userID string, filter *model.HoldingQueryFilter) ([]*model.HoldingResponse, int64, error)
	Update(ctx context.Context, id int64, userID string, dto *model.UpdateHoldingDTO) (*model.HoldingResponse, error)
	Delete(ctx context.Context, id int64, userID string) error
}

type holdingService struct {
	repo repository.HoldingRepository
}

func NewHoldingService(repo repository.HoldingRepository) HoldingService {
	return &holdingService{repo: repo}
}

func (s *holdingService) Create(ctx context.Context, userID string, dto *model.CreateHoldingDTO) (*model.HoldingResponse, error) {
	month, year := 1, 2025
	if dto.Month != nil {
		month = *dto.Month
	}
	if dto.Year != nil {
		year = *dto.Year
	}
	now := time.Now()
	h := &model.Holding{
		UserID:         userID,
		Name:           dto.Name,
		Platform:       dto.Platform,
		HoldingTypeID:  dto.HoldingTypeID,
		Currency:       dto.Currency,
		InvestedAmount: dto.InvestedAmount,
		CurrentValue:   dto.CurrentValue,
		Units:          dto.Units,
		AvgBuyPrice:    dto.AvgBuyPrice,
		CurrentPrice:   dto.CurrentPrice,
		Notes:          dto.Notes,
		Month:          month,
		Year:           year,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	created, err := s.repo.Create(ctx, h)
	if err != nil {
		return nil, err
	}
	return created.ToResponse(), nil
}

func (s *holdingService) GetByID(ctx context.Context, id int64, userID string) (*model.HoldingResponse, error) {
	h, err := s.repo.GetByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	return h.ToResponse(), nil
}

func (s *holdingService) ListByUserID(ctx context.Context, userID string, filter *model.HoldingQueryFilter) ([]*model.HoldingResponse, int64, error) {
	list, total, err := s.repo.ListByUserID(ctx, userID, filter)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*model.HoldingResponse, 0, len(list))
	for _, h := range list {
		out = append(out, h.ToResponse())
	}
	return out, total, nil
}

func (s *holdingService) Update(ctx context.Context, id int64, userID string, dto *model.UpdateHoldingDTO) (*model.HoldingResponse, error) {
	h, err := s.repo.Update(ctx, id, userID, dto)
	if err != nil {
		return nil, err
	}
	return h.ToResponse(), nil
}

func (s *holdingService) Delete(ctx context.Context, id int64, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}
