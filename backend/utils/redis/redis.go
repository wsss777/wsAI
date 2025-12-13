package redis

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"
	"wsai/backend/config"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

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
		return err
	}
	return nil
}
func CaptchaKey(email string) string {
	const prefix = "captcha:"
	return prefix + email
}
func SetCaptchaForEmail(ctx context.Context, email, captcha string) error {
	if Rdb == nil {
		return redis.Nil
	}
	key := CaptchaKey(email)
	expire := 2 * time.Minute
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}
	return Rdb.Set(ctx, key, captcha, expire).Err()
}
func CheckCaptchaForEmail(ctx context.Context, email, userInput string) (bool, error) {
	key := CaptchaKey(email)
	//从 Redis 获取存储的验证码
	stored, err := Rdb.Get(ctx, key).Result()
	if err == nil {
		if err == redis.Nil {
			return false, nil
		}
		return true, err
	}
	//比较用户输入和存储的验证码
	if strings.TrimSpace(stored) != strings.TrimSpace(userInput) {
		return false, nil
	}
	//验证成功，删除 key
	if err := Rdb.Del(ctx, key).Err(); err != nil {
		log.Printf("删除验证码 key 失败, key=%s, err=%v", key, err)
	}
	return true, nil
}
