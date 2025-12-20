package captcha

import (
	"context"
	"log"
	"strings"
	"time"
	redis2 "wsai/backend/internal/redis"

	"github.com/redis/go-redis/v9"
)

func CaptchaKey(email string) string {
	const prefix = "captcha:"
	return prefix + email
}
func SetCaptchaForEmail(ctx context.Context, email, captcha string) error {
	if redis2.Rdb == nil {
		return redis.Nil
	}
	key := CaptchaKey(email)
	expire := 2 * time.Minute
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}
	return redis2.Rdb.Set(ctx, key, captcha, expire).Err()
}
func CheckCaptchaForEmail(ctx context.Context, email, userInput string) (bool, error) {
	key := CaptchaKey(email)
	//从 Redis 获取存储的验证码
	stored, err := redis2.Rdb.Get(ctx, key).Result()
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
	if err := redis2.Rdb.Del(ctx, key).Err(); err != nil {
		log.Printf("删除验证码 key 失败, key=%s, err=%v", key, err)
	}
	return true, nil
}
