package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/port"
)

type tokenStore struct {
	client *goredis.Client
	logger *zap.Logger
}

func NewTokenStore(client *goredis.Client, logger *zap.Logger) port.TokenStore {
	return &tokenStore{client: client, logger: logger}
}

func (s *tokenStore) Store(ctx context.Context, userID string, token string, ttl time.Duration) error {
	tokenKey := tokenKey(token)
	userSetKey := userTokensKey(userID)

	pipe := s.client.TxPipeline()
	pipe.Set(ctx, tokenKey, userID, ttl)
	pipe.SAdd(ctx, userSetKey, token)
	pipe.Expire(ctx, userSetKey, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		s.logger.Error("failed to store refresh token", zap.Error(err))
		return err
	}
	return nil
}

func (s *tokenStore) GetUserID(ctx context.Context, token string) (string, error) {
	userID, err := s.client.Get(ctx, tokenKey(token)).Result()
	if errors.Is(err, goredis.Nil) {
		return "", errors.New("token not found")
	}
	if err != nil {
		s.logger.Error("failed to get userID", zap.Error(err))
		return "", err
	}
	return userID, nil
}

func (s *tokenStore) Delete(ctx context.Context, token string) error {
	tokenKey := tokenKey(token)

	userID, err := s.client.Get(ctx, tokenKey).Result()
	if err != nil {
		return err
	}

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, tokenKey)
	pipe.SRem(ctx, userTokensKey(userID), token)

	if _, err := pipe.Exec(ctx); err != nil {
		s.logger.Error("failed to delete token", zap.Error(err))
		return err
	}
	return nil
}

func (s *tokenStore) DeleteByUserID(ctx context.Context, userID string) error {
	userSetKey := userTokensKey(userID)

	tokens, err := s.client.SMembers(ctx, userSetKey).Result()
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return nil
	}

	pipe := s.client.TxPipeline()
	for _, token := range tokens {
		pipe.Del(ctx, tokenKey(token))
	}
	pipe.Del(ctx, userSetKey)

	if _, err := pipe.Exec(ctx); err != nil {
		s.logger.Error("failed to delete user tokens", zap.Error(err))
		return err
	}

	s.logger.Info("deleted all refresh tokens for user", zap.String("user_id", userID))
	return nil
}

func tokenKey(token string) string {
	return fmt.Sprintf("refresh_token:%s", token)
}

func userTokensKey(userID string) string {
	return fmt.Sprintf("user_refresh_tokens:%s", userID)
}
