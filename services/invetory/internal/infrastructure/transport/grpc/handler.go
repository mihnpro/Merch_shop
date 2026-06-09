package grpc

import (
	"context"

	inventorypb "github.com/mihnpro/Merch_shop/services/invetory/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/domain"
)

func (g *GRPCServer) CheckStock(ctx context.Context, req *inventorypb.CheckStockRequest) (*inventorypb.CheckStockResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	view, err := g.read.CheckStock(ctx, dto.CheckStockInput{
		ProductID: req.GetProductId(),
		Qty:       int(req.GetQty()),
	})
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toCheckStockProto(view), nil
}

func (g *GRPCServer) ReserveStock(ctx context.Context, req *inventorypb.ReserveStockRequest) (*inventorypb.Reservation, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	view, err := g.write.ReserveStock(ctx, toReserveStockInput(req))
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toReservationProto(view), nil
}

func (g *GRPCServer) ReleaseReserve(ctx context.Context, req *inventorypb.ReleaseReserveRequest) (*inventorypb.Reservation, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	view, err := g.write.ReleaseReserve(ctx, dto.ReleaseReserveInput{
		OrderID: req.GetOrderId(),
		Reason:  req.GetReason(),
	})
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toReservationProto(view), nil
}

func (g *GRPCServer) AdjustStock(ctx context.Context, req *inventorypb.AdjustStockRequest) (*inventorypb.AdjustStockResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	view, err := g.write.AdjustStock(ctx, dto.AdjustStockInput{
		ProductID:   req.GetProductId(),
		Delta:       int(req.GetDelta()),
		OperationID: req.GetOperationId(),
		Reason:      req.GetReason(),
	})
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toAdjustStockProto(view), nil
}
