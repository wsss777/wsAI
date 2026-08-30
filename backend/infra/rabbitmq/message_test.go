package rabbitmq

import (
	"encoding/json"
	"testing"
)

func TestGenerateMessageMQParaKeepsMessageID(t *testing.T) {
	data, err := GenerateMessageMQPara("e972d4b8-9a5e-4c78-9f94-e7b595dc8a72", "session-1", "hello", "user-1", true)
	if err != nil {
		t.Fatalf("GenerateMessageMQPara returned error: %v", err)
	}
	var payload MessageMQPara
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.MessageID != "e972d4b8-9a5e-4c78-9f94-e7b595dc8a72" {
		t.Fatalf("message ID was lost: %#v", payload)
	}
}
