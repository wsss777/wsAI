package ai

import (
	"context"
	"sync"
)

var ctx = context.Background()

// AIHelperManager 管理用户、会话与 AIHelper 的映射关系。
type AIHelperManager struct {
	helpers map[string]map[string]*AIHelper
	mu      sync.RWMutex
}

// NewAIHelperManager 创建新的管理器实例。
func NewAIHelperManager() *AIHelperManager {
	return &AIHelperManager{
		helpers: make(map[string]map[string]*AIHelper),
	}
}

// GetOrCreateAIHelper 获取或创建 AIHelper。
func (m *AIHelperManager) GetOrCreateAIHelper(username string, sessionID string, modelType string, config map[string]interface{}) (*AIHelper, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取用户的会话映射
	userHelpers, exists := m.helpers[username]
	if !exists {
		userHelpers = make(map[string]*AIHelper)
		m.helpers[username] = userHelpers
	}

	// 检查会话是否已存在
	helper, exists := userHelpers[sessionID]
	if exists && helper.GetModelType() == modelType {
		return helper, nil
	}
	// 创建新的 AIHelper。
	factory := GetGlobalFactory()
	helper, err := factory.CreateAIHelper(ctx, modelType, sessionID, config)
	if err != nil {
		return nil, err
	}
	if exists {
		helper.RestoreMessages(userHelpers[sessionID].GetAllMessage())
	}

	userHelpers[sessionID] = helper
	return helper, nil
}

// GetHelper 获取指定用户和会话的 AIHelper。
func (m *AIHelperManager) GetAIHelper(username string, sessionID string) (*AIHelper, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	userHelpers, exists := m.helpers[username]
	if !exists {
		return nil, false
	}
	helper, exists := userHelpers[sessionID]
	return helper, exists
}

// RemoveAIHelper 移除指定用户和会话的 AIHelper。
func (m *AIHelperManager) RemoveAIHelper(userName string, sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	userHelpers, exists := m.helpers[userName]
	if !exists {
		return
	}
	delete(userHelpers, sessionID)
	// 如果用户没有会话了，清理用户映射
	if len(userHelpers) == 0 {
		delete(m.helpers, userName)
	}
}

// GetUserSessions 获取指定用户的全部会话 ID。
func (m *AIHelperManager) GetUserSessions(userName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	userHelpers, exists := m.helpers[userName]
	if !exists {
		return []string{}
	}

	sessionIDs := make([]string, 0, len(userHelpers))
	// 取出所有键。
	for sessionID := range userHelpers {
		sessionIDs = append(sessionIDs, sessionID)
	}

	return sessionIDs
}

// 全局管理器实例
var globalManager *AIHelperManager
var once sync.Once

// GetGlobalManager 获取全局管理器实例
func GetGlobalManager() *AIHelperManager {
	once.Do(func() {
		globalManager = NewAIHelperManager()
	})
	return globalManager
}
