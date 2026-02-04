package repository

import (
	"context"
	"errors"
	"fmt"

	"fiberbackend/internal/model"

	"gorm.io/gorm"
)

var (
	ErrHoldingNotFound = errors.New("holding not found")
)

type HoldingRepository interface {
	Create(ctx context.Context, holding *model.Holding) (*model.Holding, error)
	GetByID(ctx context.Context, id int64) (*model.Holding, error)
	GetByIDAndUserID(ctx context.Context, id int64, userID string) (*model.Holding, error)
	ListByUserID(ctx context.Context, userID string, filter *model.HoldingQueryFilter) ([]*model.Holding, int64, error)
	Update(ctx context.Context, id int64, userID string, dto *model.UpdateHoldingDTO) (*model.Holding, error)
	Delete(ctx context.Context, id int64, userID string) error
}

type holdingRepository struct {
	db *gorm.DB
}

func NewHoldingRepository(db *gorm.DB) HoldingRepository {
	return &holdingRepository{db: db}
}

func (r *holdingRepository) Create(ctx context.Context, holding *model.Holding) (*model.Holding, error) {
	if err := r.db.WithContext(ctx).Create(holding).Error; err != nil {
		return nil, fmt.Errorf("create holding: %w", err)
	}
	return holding, nil
}

func (r *holdingRepository) GetByID(ctx context.Context, id int64) (*model.Holding, error) {
	var h model.Holding
	err := r.db.WithContext(ctx).Preload("HoldingType").First(&h, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHoldingNotFound
		}
		return nil, fmt.Errorf("get holding: %w", err)
	}
	return &h, nil
}

func (r *holdingRepository) GetByIDAndUserID(ctx context.Context, id int64, userID string) (*model.Holding, error) {
	var h model.Holding
	err := r.db.WithContext(ctx).Preload("HoldingType").First(&h, "id = ? AND user_id = ?", id, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHoldingNotFound
		}
		return nil, fmt.Errorf("get holding: %w", err)
	}
	return &h, nil
}

func (r *holdingRepository) ListByUserID(ctx context.Context, userID string, filter *model.HoldingQueryFilter) ([]*model.Holding, int64, error) {
	if filter == nil {
		filter = &model.HoldingQueryFilter{Limit: 20, Offset: 0}
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	query := r.db.WithContext(ctx).Model(&model.Holding{}).Where("user_id = ?", userID)
	if filter.Month != nil {
		query = query.Where("month = ?", *filter.Month)
	}
	if filter.Year != nil {
		query = query.Where("year = ?", *filter.Year)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count holdings: %w", err)
	}

	var list []*model.Holding
	err := query.Preload("HoldingType").
		Order("year DESC, month DESC, id DESC").
		Limit(filter.Limit).Offset(filter.Offset).
		Find(&list).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list holdings: %w", err)
	}
	return list, total, nil
}

func (r *holdingRepository) Update(ctx context.Context, id int64, userID string, dto *model.UpdateHoldingDTO) (*model.Holding, error) {
	updates := make(map[string]interface{})
	if dto.Name != nil {
		updates["name"] = *dto.Name
	}
	if dto.Platform != nil {
		updates["platform"] = *dto.Platform
	}
	if dto.HoldingTypeID != nil {
		updates["holding_type_id"] = *dto.HoldingTypeID
	}
	if dto.Currency != nil {
		updates["currency"] = *dto.Currency
	}
	if dto.InvestedAmount != nil {
		updates["invested_amount"] = *dto.InvestedAmount
	}
	if dto.CurrentValue != nil {
		updates["current_value"] = *dto.CurrentValue
	}
	if dto.Units != nil {
		updates["units"] = *dto.Units
	}
	if dto.AvgBuyPrice != nil {
		updates["avg_buy_price"] = *dto.AvgBuyPrice
	}
	if dto.CurrentPrice != nil {
		updates["current_price"] = *dto.CurrentPrice
	}
	if dto.Notes != nil {
		updates["notes"] = *dto.Notes
	}
	if dto.Month != nil {
		updates["month"] = *dto.Month
	}
	if dto.Year != nil {
		updates["year"] = *dto.Year
	}

	if len(updates) == 0 {
		return r.GetByIDAndUserID(ctx, id, userID)
	}

	res := r.db.WithContext(ctx).Model(&model.Holding{}).Where("id = ? AND user_id = ?", id, userID).Updates(updates)
	if res.Error != nil {
		return nil, fmt.Errorf("update holding: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrHoldingNotFound
	}
	return r.GetByIDAndUserID(ctx, id, userID)
}

func (r *holdingRepository) Delete(ctx context.Context, id int64, userID string) error {
	res := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.Holding{})
	if res.Error != nil {
		return fmt.Errorf("delete holding: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrHoldingNotFound
	}
	return nil
}
