package session

import "context"

// Repository 定义应用服务使用的会话持久化边界。
type Repository interface {
	FindByID(ctx context.Context, id string) (*Entity, error)
}
