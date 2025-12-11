package user

import (
	"context"
	"github.com/cloudwego/eino/schema"
	"wsai/backend/internal/model"
)

func StreamChat(ctx context.Context, userMsg string) (model.ChatStream, error) {
	return model.Get().Stream(ctx, []*schema.Message{
		{Role: "user", Content: userMsg},
	})
}
