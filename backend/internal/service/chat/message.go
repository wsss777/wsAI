package chat

import (
	"encoding/json"
	"time"
	"wsai/backend/internal/logger"
	"wsai/backend/internal/model"
	"wsai/backend/internal/repository/message"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// MessageMQParam 定义投递到 RabbitMQ 的消息参数结构
type MessageMQPara struct {
	SessionID string `json:"session_id"`
	Content   string `json:"content"`
	Username  string `json:"username"`
	IsUser    bool   `json:"is_user"`
}

// GenerateMessageMQParam 生成要发送到 MQ 的 JSON 字节
func GenerateMessageMQPara(sessionID string, content string, username string, isUser bool) ([]byte, error) {
	para := MessageMQPara{
		SessionID: sessionID,
		Content:   content,
		Username:  username,
		IsUser:    isUser,
	}

	data, err := json.Marshal(para)
	if err != nil {
		logger.L().Error("RabbitMQ message marshal failed in generateMessageMQPara",
			zap.Error(err),
			zap.String("sessionID", sessionID),
			zap.String("username", username),
			zap.Bool("isUser", isUser))
		return nil, err
	}
	return data, nil
}

// ProcessMessageDelivery 处理 RabbitMQ 投递的消息
func ProcessMessageDelivery(msg amqp.Delivery) error {
	var para MessageMQPara
	if err := json.Unmarshal(msg.Body, &para); err != nil {
		logger.L().Error("RabbitMQ message unmarshal failed in processMessageDelivery",
			zap.Error(err),
		)
		return err
	}
	logger.L().Info("RabbitMQ received message",
		zap.String("session_id", para.SessionID),
		zap.String("user_name", para.Username),
		zap.Bool("is_user", para.IsUser),
		zap.Int("content_length", len(para.Content)),
	)

	newMsg := &model.Message{
		SessionID: para.SessionID,
		Content:   para.Content,
		UserName:  para.Username,
		IsUser:    para.IsUser,
		CreatedAt: time.Now(),
	}
	go func(m *model.Message, tag uint64) {
		if _, err := message.CreateMessage(m); err != nil {
			logger.L().Error("Async save chat message to DB failed",
				zap.Error(err),
				zap.String("session_id", m.SessionID),
				zap.Uint64("delivery_tag", tag),
				zap.String("username", m.UserName),
			)
		} else {
			logger.L().Debug("Chat message saved to DB successfully",
				zap.String("session_id", m.SessionID),
				zap.Uint64("delivery_tag", tag),
			)
		}
	}(newMsg, msg.DeliveryTag)
	return nil
}
