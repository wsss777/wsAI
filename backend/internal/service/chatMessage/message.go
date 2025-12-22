package chatMessage

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
func ProcessMessageDelivery(msg *amqp.Delivery) error {
	var para MessageMQPara
	if err := json.Unmarshal(msg.Body, &para); err != nil {
		logger.L().Error("RabbitMQ message unmarshal failed in processMessageDelivery",
			zap.Error(err),
			zap.Uint64("delivery_tag", msg.DeliveryTag),
		)
		msg.Nack(false, true)
		return err
	}
	logger.L().Info("RabbitMQ received message",
		zap.String("session_id", para.SessionID),
		zap.String("user_name", para.Username),
		zap.Bool("is_user", para.IsUser),
		zap.Int("content_length", len(para.Content)),
		zap.Uint64("delivery_tag", msg.DeliveryTag),
	)

	newMsg := &model.Message{
		SessionID: para.SessionID,
		Content:   para.Content,
		UserName:  para.Username,
		IsUser:    para.IsUser,
		CreatedAt: time.Now(),
	}
	if _, err := message.CreateMessage(newMsg); err != nil {
		logger.L().Error("Save chatMessage message to DB failed",
			zap.Error(err),
			zap.String("session_id", newMsg.SessionID),
			zap.Uint64("delivery_tag", msg.DeliveryTag),
			zap.String("username", newMsg.UserName),
		)
		// 保存失败：拒绝消息并 requeue（让同一个或别的消费者重试）
		msg.Nack(false, true)
		return err
	}
	logger.L().Debug("Chat message saved to DB successfully",
		zap.String("session_id", newMsg.SessionID),
		zap.Uint64("delivery_tag", msg.DeliveryTag),
	)
	msg.Ack(false)
	return nil
}
