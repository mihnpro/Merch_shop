package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	userpb "github.com/mihnpro/Merch_shop/services/user_customer/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/dto"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/domain"
)

const metaUserID = "x-user-id"

func userIDFromMeta(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", domain.ErrInvalidToken
	}
	vals := md.Get(metaUserID)
	if len(vals) == 0 || vals[0] == "" {
		return "", domain.ErrInvalidToken
	}
	return vals[0], nil
}


func (g *GRPCServer) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.AuthResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	view, err := g.auth.Register(ctx, toRegisterInput(req))
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return &userpb.AuthResponse{User: toUserProto(view)}, nil
}

func (g *GRPCServer) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.AuthResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	result, err := g.auth.Login(ctx, toLoginInput(req))
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return &userpb.AuthResponse{
		User:   toUserProto(result.User),
		Tokens: toTokenPairProto(result.Tokens),
	}, nil
}

func (g *GRPCServer) Refresh(ctx context.Context, req *userpb.RefreshRequest) (*userpb.TokenPair, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	tokens, err := g.auth.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toTokenPairProto(tokens), nil
}

func (g *GRPCServer) Logout(ctx context.Context, req *userpb.LogoutRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	if err := g.auth.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, NewGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}


func (g *GRPCServer) GetUser(ctx context.Context, _ *emptypb.Empty) (*userpb.User, error) {
	userID, err := userIDFromMeta(ctx)
	if err != nil {
		return nil, NewGRPCError(err)
	}
	view, err := g.user.GetUser(ctx, userID)
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toUserProto(view), nil
}

func (g *GRPCServer) ChangePassword(ctx context.Context, req *userpb.ChangePasswordRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	userID, err := userIDFromMeta(ctx)
	if err != nil {
		return nil, NewGRPCError(err)
	}
	if err := g.user.ChangePassword(ctx, dto.ChangePasswordInput{
		UserID:      userID,
		OldPassword: req.GetOldPassword(),
		NewPassword: req.GetNewPassword(),
	}); err != nil {
		return nil, NewGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (g *GRPCServer) GetBalance(ctx context.Context, _ *emptypb.Empty) (*userpb.Balance, error) {
	userID, err := userIDFromMeta(ctx)
	if err != nil {
		return nil, NewGRPCError(err)
	}
	bal, err := g.user.GetBalance(ctx, userID)
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toBalanceProto(bal), nil
}

func (g *GRPCServer) DeductPoints(ctx context.Context, req *userpb.DeductPointsRequest) (*userpb.Balance, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	bal, err := g.user.DeductPoints(ctx, dto.DeductPointsInput{
		UserID:      req.GetUserId(),
		Amount:      req.GetAmount(),
		OperationID: req.GetOperationId(),
		OrderID:     req.GetOrderId(),
		Reason:      req.GetReason(),
	})
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toBalanceProto(bal), nil
}

func (g *GRPCServer) AddPoints(ctx context.Context, req *userpb.AddPointsRequest) (*userpb.Balance, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	bal, err := g.user.AddPoints(ctx, dto.AddPointsInput{
		UserID:      req.GetUserId(),
		Amount:      req.GetAmount(),
		OperationID: req.GetOperationId(),
		OrderID:     req.GetOrderId(),
		Reason:      req.GetReason(),
	})
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toBalanceProto(bal), nil
}

func (g *GRPCServer) GrantPoints(ctx context.Context, req *userpb.GrantPointsRequest) (*userpb.Balance, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	bal, err := g.admin.GrantPoints(ctx, dto.GrantPointsInput{
		UserID:      req.GetUserId(),
		Amount:      req.GetAmount(),
		OperationID: req.GetOperationId(),
		Reason:      req.GetReason(),
	})
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toBalanceProto(bal), nil
}

func (g *GRPCServer) ResetPassword(ctx context.Context, req *userpb.ResetPasswordRequest) (*userpb.ResetPasswordResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	newPass, err := g.admin.ResetPassword(ctx, req.GetUserId())
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return &userpb.ResetPasswordResponse{NewPassword: newPass}, nil
}

func (g *GRPCServer) BlockUser(ctx context.Context, req *userpb.BlockUserRequest) (*userpb.User, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	blocked := req.GetNewStatus() == userpb.UserStatus_USER_STATUS_BLOCKED
	view, err := g.admin.BlockUser(ctx, req.GetUserId(), blocked)
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toUserProto(view), nil
}

func (g *GRPCServer) ChangeRole(ctx context.Context, req *userpb.ChangeRoleRequest) (*userpb.User, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	view, err := g.admin.ChangeRole(ctx, req.GetUserId(), req.GetRoleCode())
	if err != nil {
		return nil, NewGRPCError(err)
	}
	return toUserProto(view), nil
}

func (g *GRPCServer) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	views, nextToken, err := g.admin.ListUsers(ctx, dto.ListUsersInput{
		Search:    req.GetSearch(),
		RoleCode:  req.GetRoleCode(),
		Status:    toUserStatusString(req.GetStatus()),
		PageSize:  int(req.GetPageSize()),
		PageToken: req.GetPageToken(),
	})
	if err != nil {
		return nil, NewGRPCError(err)
	}
	users := make([]*userpb.User, 0, len(views))
	for _, v := range views {
		users = append(users, toUserProto(v))
	}
	return &userpb.ListUsersResponse{Users: users, NextPageToken: nextToken}, nil
}

func (g *GRPCServer) GetTransactions(ctx context.Context, req *userpb.GetTransactionsRequest) (*userpb.GetTransactionsResponse, error) {
	if req == nil {
		return nil, NewGRPCError(domain.ErrEmptyRequest)
	}
	userID := req.GetUserId()
	if userID == "" {
		var err error
		userID, err = userIDFromMeta(ctx)
		if err != nil {
			return nil, NewGRPCError(err)
		}
	}
	txs, nextToken, err := g.admin.GetTransactions(ctx, dto.GetTransactionsInput{
		UserID:    userID,
		PageSize:  int(req.GetPageSize()),
		PageToken: req.GetPageToken(),
	})
	if err != nil {
		return nil, NewGRPCError(err)
	}
	protoTxs := make([]*userpb.PointsTransaction, 0, len(txs))
	for _, t := range txs {
		protoTxs = append(protoTxs, toTransactionProto(t))
	}
	return &userpb.GetTransactionsResponse{Transactions: protoTxs, NextPageToken: nextToken}, nil
}
