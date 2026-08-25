package ai

import (
	"fmt"
	"testing"
	"time"
)

func TestAIHelperKeepsOnlyRecentContextMessages(t *testing.T) {
	helper := NewAIHelper(nil, "session-1")

	for i := 0; i < MaxContextMessages+2; i++ {
		helper.AddMessage(fmt.Sprintf("message-%d", i), "alice", i%2 == 0, false)
	}

	messages := helper.GetAllMessage()
	if len(messages) != MaxContextMessages {
		t.Fatalf("context message count = %d, want %d", len(messages), MaxContextMessages)
	}
	if !messages[0].IsUser {
		t.Fatal("context must begin with a user message")
	}
	if messages[0].Content != "message-2" {
		t.Fatalf("oldest retained message = %q, want %q", messages[0].Content, "message-2")
	}
}

func TestAIHelperManagerRemovesExpiredHelpers(t *testing.T) {
	manager := NewAIHelperManager()
	manager.helpers["alice"] = map[string]*managedHelper{
		"session-1": {
			helper:     NewAIHelper(nil, "session-1"),
			lastAccess: time.Now().Add(-helperIdleTTL),
		},
	}

	manager.removeExpiredLocked(time.Now())
	if _, exists := manager.helpers["alice"]; exists {
		t.Fatal("expired helper was not removed")
	}
}
