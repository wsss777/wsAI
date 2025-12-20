package redis

import (
	"context"
	"strconv"
	"time"
	"wsai/backend/config"
	"wsai/backend/internal/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Rdb 全局 Redis 客户端
var Rdb *redis.Client

// Init 初始化 Redis 连接
func Init() error {
	host := config.C.RedisConfig.Host
	port := config.C.RedisConfig.Port
	password := config.C.RedisConfig.Password
	db := config.C.RedisConfig.DB

	addr := host + ":" + strconv.Itoa(port)

	Rdb = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     100,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := Rdb.Ping(ctx).Err()
	if err != nil {
		logger.L().Error("Redis 连接失败",
			zap.Error(err),
			zap.String("addr", addr),
			zap.Int("db", db),
		)
		return err
	}

	logger.L().Info("Redis 连接成功",
		zap.String("addr", addr),
		zap.Int("db", db),
	)
	return nil
}

func Close() error {
	if Rdb != nil {
		return Rdb.Close()
	}
	return nil
}
