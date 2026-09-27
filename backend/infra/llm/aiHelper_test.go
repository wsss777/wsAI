package ai

import (
	"testing"
	"time"
)

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
