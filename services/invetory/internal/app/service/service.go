package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain/repository"
	domainsvc "github.com/mihnpro/Merch_shop/services/invetory/internal/domain/service"
)

type InventoryService interface {
	ReserveStock(ctx context.Context, in dto.ReserveStockInput) (dto.ReservationView, error)
	ReleaseReserve(ctx context.Context, in dto.ReleaseReserveInput) (dto.ReservationView, error)
	AdjustStock(ctx context.Context, in dto.AdjustStockInput) (dto.StockView, error)
}

type inventoryService struct {
	stock              repository.StockRepository
	reservations       repository.ReservationRepository
	stockFactory       *domainsvc.StockFactory
	reservationFactory *domainsvc.ReservationFactory
}

func NewInventoryService(stock repository.StockRepository, reservations repository.ReservationRepository) InventoryService {
	return &inventoryService{
		stock:              stock,
		reservations:       reservations,
		stockFactory:       domainsvc.NewStockFactory(),
		reservationFactory: domainsvc.NewReservationFactory(),
	}
}


func (s *inventoryService) ReserveStock(ctx context.Context, in dto.ReserveStockInput) (dto.ReservationView, error) {
	orderID, err := parseID(in.OrderID, domain.ErrInvalidOrderID)
	if err != nil {
		return dto.ReservationView{}, err
	}

	items := make([]domainsvc.ReservationItemInput, 0, len(in.Items))
	for _, raw := range in.Items {
		productID, err := parseID(raw.ProductID, domain.ErrInvalidProductID)
		if err != nil {
			return dto.ReservationView{}, err
		}
		items = append(items, domainsvc.ReservationItemInput{
			ProductID: productID,
			Qty:       raw.Qty,
		})
	}

	reservation, err := s.reservationFactory.Create(domainsvc.ReserveInput{
		OrderID: orderID,
		Items:   items,
	})
	if err != nil {
		return dto.ReservationView{}, err
	}

	saved, err := s.reservations.Reserve(ctx, reservation)
	if err != nil {
		return dto.ReservationView{}, err
	}
	return dto.ToReservationView(saved), nil
}

func (s *inventoryService) ReleaseReserve(ctx context.Context, in dto.ReleaseReserveInput) (dto.ReservationView, error) {
	orderID, err := parseID(in.OrderID, domain.ErrInvalidOrderID)
	if err != nil {
		return dto.ReservationView{}, err
	}

	released, err := s.reservations.Release(ctx, orderID, in.Reason)
	if err != nil {
		return dto.ReservationView{}, err
	}
	return dto.ToReservationView(released), nil
}

func (s *inventoryService) AdjustStock(ctx context.Context, in dto.AdjustStockInput) (dto.StockView, error) {
	productID, err := parseID(in.ProductID, domain.ErrInvalidProductID)
	if err != nil {
		return dto.StockView{}, err
	}
	operationID, err := parseID(in.OperationID, domain.ErrInvalidOperationID)
	if err != nil {
		return dto.StockView{}, err
	}

	adjustment, err := s.stockFactory.NewAdjustment(domainsvc.AdjustStockInput{
		OperationID: operationID,
		ProductID:   productID,
		Delta:       in.Delta,
		Reason:      in.Reason,
	})
	if err != nil {
		return dto.StockView{}, err
	}

	stock, err := s.stock.Adjust(ctx, adjustment)
	if err != nil {
		return dto.StockView{}, err
	}
	return dto.ToStockView(stock), nil
}

func parseID(raw string, invalid error) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, invalid
	}
	return id, nil
}
