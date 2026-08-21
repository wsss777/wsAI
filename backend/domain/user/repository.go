package user

import "context"

// Repository 定义应用服务使用的用户持久化边界。
type Repository interface {
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
