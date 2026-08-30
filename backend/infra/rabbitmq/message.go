package rabbitmq

import (
	"encoding/json"
	"time"
	"wsai/backend/infra/logger"
	message "wsai/backend/infra/mysql/repository"
	"wsai/backend/model"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// MessageMQPara 定义投递到 RabbitMQ 的消息参数结构
type MessageMQPara struct {
	MessageID string `json:"message_id"`
	SessionID string `json:"session_id"`
	Content   string `json:"content"`
	Username  string `json:"username"`
	IsUser    bool   `json:"is_user"`
}

// GenerateMessageMQPara 生成要发送到消息队列的 JSON 字节。
func GenerateMessageMQPara(messageID string, sessionID string, content string, username string, isUser bool) ([]byte, error) {
	para := MessageMQPara{
		MessageID: messageID,
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
func ProcessMessageDelivery(msg *amqp.Delivery) error {
	var para MessageMQPara
	if err := json.Unmarshal(msg.Body, &para); err != nil {
		logger.L().Error("RabbitMQ message unmarshal failed in processMessageDelivery",
			zap.Error(err),
			zap.Uint64("delivery_tag", msg.DeliveryTag),
		)
		return err
	}
	logger.L().Info("RabbitMQ received message",
		zap.String("message_id", para.MessageID),
		zap.String("session_id", para.SessionID),
		zap.String("user_name", para.Username),
		zap.Bool("is_user", para.IsUser),
		zap.Int("content_length", len(para.Content)),
		zap.Uint64("delivery_tag", msg.DeliveryTag),
	)

	newMsg := &model.Message{
		MessageID: para.MessageID,
		SessionID: para.SessionID,
		Content:   para.Content,
		UserName:  para.Username,
		IsUser:    para.IsUser,
		CreatedAt: time.Now(),
	}
	created, err := message.CreateMessageIfAbsent(newMsg)
	if err != nil {
		logger.L().Error("Save chatMessage message to DB failed",
			zap.Error(err),
			zap.String("message_id", newMsg.MessageID),
			zap.String("session_id", newMsg.SessionID),
			zap.Uint64("delivery_tag", msg.DeliveryTag),
			zap.String("username", newMsg.UserName),
		)
		return err
	}
	if !created {
		logger.L().Info("RabbitMQ duplicate message ignored",
			zap.String("message_id", newMsg.MessageID),
			zap.Uint64("delivery_tag", msg.DeliveryTag),
		)
		return nil
	}
	logger.L().Debug("Chat message saved to DB successfully",
		zap.String("message_id", newMsg.MessageID),
		zap.String("session_id", newMsg.SessionID),
		zap.Uint64("delivery_tag", msg.DeliveryTag),
	)
	return nil
}
