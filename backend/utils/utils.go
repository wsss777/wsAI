package utils

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"math/rand"
	"wsai/backend/model"

	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// VerifyPassword 同时兼容历史 MD5，成功登录后调用方应升级为 bcrypt。
func VerifyPassword(storedHash, password string) (matched bool, needsUpgrade bool) {
	if len(storedHash) == 32 {
		return subtle.ConstantTimeCompare([]byte(storedHash), []byte(MD5(password))) == 1, true
	}
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) == nil, false
}
func GenerateUUID() string {
	return uuid.New().String()
}

// ConvertToModelMessage 将模式消息转换为数据库存储格式。
func ConvertToModelMessage(sessionID string, username string, msg *schema.Message) *model.Message {
	return &model.Message{
		SessionID: sessionID,
		UserName:  username,
		Content:   msg.Content,
	}
}

// ConvertToSchemaMessages 将数据库存储格式转换为模式消息。
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
