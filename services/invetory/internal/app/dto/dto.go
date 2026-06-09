package dto

import "time"


type ReserveItemInput struct {
	ProductID string
	Qty       int
}

type ReserveStockInput struct {
	OrderID string
	Items   []ReserveItemInput
}

type ReleaseReserveInput struct {
	OrderID string
	Reason  string
}

type AdjustStockInput struct {
	ProductID   string
	Delta       int
	OperationID string
	Reason      string
}


type CheckStockInput struct {
	ProductID string
	Qty       int
}


type StockView struct {
	ProductID string
	Available int
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CheckStockView struct {
	ProductID string
	Available int
	InStock   bool
}

type ReservationItemView struct {
	ProductID string
	Qty       int
}

type ReservationView struct {
	ID             string
	OrderID        string
	Items          []ReservationItemView
	ReleasedReason string
	ReleasedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
