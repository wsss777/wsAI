package user

import (
	"context"
	"wsai/backend/internal/model"

	"github.com/cloudwego/eino/schema"
)

func StreamChat(ctx context.Context, userMsg string) (model.ChatStream, error) {
	return model.Get().Stream(ctx, []*schema.Message{
		{Role: "user", Content: userMsg},
	})
}
