package jwt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
	redisclient "wsai/backend/internal/common/redis"
)

const blacklistKeyPrefix = "jwt:blacklist:"

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

func IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	if redisclient.Rdb == nil {
		return false, nil
	}

	exists, err := redisclient.Rdb.Exists(ctx, blacklistKey(token)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func blacklistKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return blacklistKeyPrefix + hex.EncodeToString(sum[:])
}
