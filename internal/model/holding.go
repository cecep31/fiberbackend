package model

import (
	"time"
)

// HoldingType is the referenced lookup table for holding_type_id.
// Table holding_types must exist (e.g. from Drizzle/migrations).
type HoldingType struct {
	ID   int16  `json:"id" gorm:"type:smallint;primaryKey"`
	Name string `json:"name" gorm:"type:text"`
}

func (HoldingType) TableName() string {
	return "holding_types"
}

// Holding represents a user's holding (investment/asset) per month/year.
type Holding struct {
	ID             int64      `json:"id" gorm:"type:bigint;primaryKey;autoIncrement"`
	UserID         string     `json:"user_id" gorm:"type:uuid;not null;index:idx_holdings_user"`
	Name           string     `json:"name" gorm:"type:text;not null"`
	Platform       string     `json:"platform" gorm:"type:text;not null"`
	HoldingTypeID  int16      `json:"holding_type_id" gorm:"type:smallint;not null;index:idx_holdings_holding_type_id"`
	Currency       string     `json:"currency" gorm:"type:char(3);not null"`
	InvestedAmount float64    `json:"invested_amount" gorm:"type:numeric(18,2);default:0;not null"`
	CurrentValue   float64    `json:"current_value" gorm:"type:numeric(18,2);default:0;not null"`
	Units          *float64   `json:"units,omitempty" gorm:"type:numeric(24,10)"`
	AvgBuyPrice    *float64   `json:"avg_buy_price,omitempty" gorm:"column:avg_buy_price;type:numeric(18,8)"`
	CurrentPrice   *float64   `json:"current_price,omitempty" gorm:"column:current_price;type:numeric(18,8)"`
	LastUpdated    *time.Time `json:"last_updated,omitempty" gorm:"type:timestamptz"`
	Notes          *string    `json:"notes,omitempty" gorm:"type:text"`
	CreatedAt      time.Time  `json:"created_at" gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"type:timestamptz;default:now();not null"`
	Month          int        `json:"month" gorm:"default:1;not null"`
	Year           int        `json:"year" gorm:"default:2025;not null"`

	User        *User        `json:"-" gorm:"foreignKey:UserID"`
	HoldingType *HoldingType `json:"holding_type,omitempty" gorm:"foreignKey:HoldingTypeID"`
}

func (Holding) TableName() string {
	return "holdings"
}

// HoldingResponse for API output.
type HoldingResponse struct {
	ID             int64        `json:"id"`
	UserID         string       `json:"user_id"`
	Name           string       `json:"name"`
	Platform       string       `json:"platform"`
	HoldingTypeID  int16        `json:"holding_type_id"`
	Currency       string       `json:"currency"`
	InvestedAmount float64      `json:"invested_amount"`
	CurrentValue   float64      `json:"current_value"`
	Units          *float64     `json:"units,omitempty"`
	AvgBuyPrice    *float64     `json:"avg_buy_price,omitempty"`
	CurrentPrice   *float64     `json:"current_price,omitempty"`
	LastUpdated    *time.Time   `json:"last_updated,omitempty"`
	Notes          *string      `json:"notes,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	Month          int          `json:"month"`
	Year           int          `json:"year"`
	HoldingType    *HoldingType `json:"holding_type,omitempty"`
}

func (h *Holding) ToResponse() *HoldingResponse {
	resp := &HoldingResponse{
		ID:             h.ID,
		UserID:         h.UserID,
		Name:           h.Name,
		Platform:       h.Platform,
		HoldingTypeID:  h.HoldingTypeID,
		Currency:       h.Currency,
		InvestedAmount: h.InvestedAmount,
		CurrentValue:   h.CurrentValue,
		Units:          h.Units,
		AvgBuyPrice:    h.AvgBuyPrice,
		CurrentPrice:   h.CurrentPrice,
		LastUpdated:    h.LastUpdated,
		Notes:          h.Notes,
		CreatedAt:      h.CreatedAt,
		UpdatedAt:      h.UpdatedAt,
		Month:          h.Month,
		Year:           h.Year,
	}
	if h.HoldingType != nil {
		resp.HoldingType = h.HoldingType
	}
	return resp
}

// CreateHoldingDTO for creating a holding.
type CreateHoldingDTO struct {
	Name           string   `json:"name" validate:"required,min=1"`
	Platform       string   `json:"platform" validate:"required,min=1"`
	HoldingTypeID  int16    `json:"holding_type_id" validate:"required"`
	Currency       string   `json:"currency" validate:"required,len=3"`
	InvestedAmount float64  `json:"invested_amount" validate:"gte=0"`
	CurrentValue   float64  `json:"current_value" validate:"gte=0"`
	Units          *float64 `json:"units,omitempty"`
	AvgBuyPrice    *float64 `json:"avg_buy_price,omitempty"`
	CurrentPrice   *float64 `json:"current_price,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
	Month          *int     `json:"month,omitempty" validate:"omitempty,gte=1,lte=12"`
	Year           *int     `json:"year,omitempty" validate:"omitempty,gte=2000"`
}

// UpdateHoldingDTO for partial update.
type UpdateHoldingDTO struct {
	Name           *string  `json:"name,omitempty" validate:"omitempty,min=1"`
	Platform       *string  `json:"platform,omitempty" validate:"omitempty,min=1"`
	HoldingTypeID  *int16   `json:"holding_type_id,omitempty"`
	Currency       *string  `json:"currency,omitempty" validate:"omitempty,len=3"`
	InvestedAmount *float64 `json:"invested_amount,omitempty" validate:"omitempty,gte=0"`
	CurrentValue   *float64 `json:"current_value,omitempty" validate:"omitempty,gte=0"`
	Units          *float64 `json:"units,omitempty"`
	AvgBuyPrice    *float64 `json:"avg_buy_price,omitempty"`
	CurrentPrice   *float64 `json:"current_price,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
	Month          *int     `json:"month,omitempty" validate:"omitempty,gte=1,lte=12"`
	Year           *int     `json:"year,omitempty" validate:"omitempty,gte=2000"`
}

// HoldingQueryFilter for list filters.
type HoldingQueryFilter struct {
	Month  *int `json:"month" query:"month"`
	Year   *int `json:"year" query:"year"`
	Limit  int  `json:"limit" query:"limit"`
	Offset int  `json:"offset" query:"offset"`
}
