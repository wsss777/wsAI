package ai

import (
	"context"
	"sync"
)

var ctx = context.Background()

// AIHelperManagerAI 助手管理器，管理用户-会话-AIHelper的映射关系
type AIHelperManager struct {
	helpers map[string]map[string]*AIHelper
	mu      sync.RWMutex
}

// NewAIHelperManager 创建新的管理器实例
func NewAIHelperManager() *AIHelperManager {
	return &AIHelperManager{
		helpers: make(map[string]map[string]*AIHelper),
	}
}

// 获取或创建AIHelper
func (m *AIHelperManager) GetOrCreateAIHelper(username string, sessionID string, modelType string, config map[string]interface{}) (*AIHelper, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	//获取用户的会话映射
	userHelpers, exists := m.helpers[username]
	if !exists {
		userHelpers = make(map[string]*AIHelper)
		m.helpers[username] = userHelpers
	}

	//检查会话是否已经存在
	helper, exists := userHelpers[sessionID]
	if exists {
		return helper, nil
	}

	//创建新的AIHelper
	factory := GetGlobal

}
