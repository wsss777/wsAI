package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"wsai/backend/internal/model"

	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

func GetRandomNumbers(num int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var builder strings.Builder

	builder.Grow(num)

	for i := 0; i < num; i++ {
		digit := r.Intn(10)
		builder.WriteByte(byte('0' + digit))
	}

	return builder.String()
}

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
func GenerateUUID() string {
	return uuid.New().String()
}

// ConvertToModelMessage 将schema消息转换为数据库存储的格式
func ConvertToModelMessage(sessionID string, username string, msg *schema.Message) *model.Message {
	return &model.Message{
		SessionID: sessionID,
		UserName:  username,
		Content:   msg.Content,
	}
}

// ConvertToSchemaMessages 将数据库存储的格式转换为schema
func ConvertToSchemaMessages(msgs []*model.Message) []*schema.Message {
	schemaMsgs := make([]*schema.Message, 0, len(msgs))
	for _, m := range msgs {
		role := schema.Assistant
		if m.IsUser {
			role = schema.User
		}
		schemaMsgs = append(schemaMsgs, &schema.Message{
			Role:    role,
			Content: m.Content,
		})
	}
	return schemaMsgs
}
