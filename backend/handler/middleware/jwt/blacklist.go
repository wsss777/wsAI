package jwt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
	redisclient "wsai/backend/infra/redis"
)

const blacklistKeyPrefix = "jwt:blacklist:"
const refreshKeyPrefix = "jwt:refresh:"

func AddTokenToBlacklist(ctx context.Context, token string, expireAt time.Time) error {
	if redisclient.Rdb == nil {
		return errors.New("Redis 客户端未初始化")
	}

	ttl := time.Until(expireAt)
	if ttl <= 0 {
		return nil
	}

	return redisclient.Rdb.Set(ctx, blacklistKey(token), "1", ttl).Err()
}

func StoreRefreshToken(ctx context.Context, token string, expireAt time.Time) error {
	if redisclient.Rdb == nil {
		return errors.New("Redis 客户端未初始化")
	}
	ttl := time.Until(expireAt)
	if ttl <= 0 {
		return errors.New("refresh token 已过期")
	}
	return redisclient.Rdb.Set(ctx, refreshKey(token), "1", ttl).Err()
}

func ConsumeRefreshToken(ctx context.Context, token string) (bool, error) {
	if redisclient.Rdb == nil {
		return false, errors.New("Redis 客户端未初始化")
	}
	value, err := redisclient.Rdb.GetDel(ctx, refreshKey(token)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return value == "1", nil
}

func RevokeRefreshToken(ctx context.Context, token string) error {
	if redisclient.Rdb == nil {
		return errors.New("Redis 客户端未初始化")
	}
	return redisclient.Rdb.Del(ctx, refreshKey(token)).Err()
}

func IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	if redisclient.Rdb == nil {
		return false, errors.New("Redis 客户端未初始化")
	}

	exists, err := redisclient.Rdb.Exists(ctx, blacklistKey(token)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func blacklistKey(token string) string {
	return blacklistKeyPrefix + tokenHash(token)
}
func refreshKey(token string) string { return refreshKeyPrefix + tokenHash(token) }
func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
