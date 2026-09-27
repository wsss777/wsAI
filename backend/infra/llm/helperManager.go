package ai

import (
	"context"
	"sync"
	"time"
)

var ctx = context.Background()

const (
	helperIdleTTL      = 30 * time.Minute
	helperCleanupEvery = time.Minute
)

// AIHelperManager 管理用户、会话与 AIHelper 的映射关系。
type AIHelperManager struct {
	helpers map[string]map[string]*managedHelper
	mu      sync.RWMutex
}

type managedHelper struct {
	helper     *AIHelper
	lastAccess time.Time
}

// NewAIHelperManager 创建新的管理器实例。
func NewAIHelperManager() *AIHelperManager {
	return &AIHelperManager{
		helpers: make(map[string]map[string]*managedHelper),
	}
}

// GetOrCreateAIHelper 获取或创建 AIHelper。
func (m *AIHelperManager) GetOrCreateAIHelper(username string, sessionID string, modelType string, config map[string]interface{}) (*AIHelper, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeExpiredLocked(time.Now())

	// 获取用户的会话映射
	userHelpers, exists := m.helpers[username]
	if !exists {
		userHelpers = make(map[string]*managedHelper)
		m.helpers[username] = userHelpers
	}

	// 检查会话是否已存在
	entry, exists := userHelpers[sessionID]
	if exists && entry.helper.GetModelType() == modelType {
		entry.lastAccess = time.Now()
		return entry.helper, nil
	}
	// 创建新的 AIHelper。
	factory := GetGlobalFactory()
	helper, err := factory.CreateAIHelper(ctx, modelType, sessionID, config)
	if err != nil {
		return nil, err
	}
	userHelpers[sessionID] = &managedHelper{helper: helper, lastAccess: time.Now()}
	return helper, nil
}

// GetHelper 获取指定用户和会话的 AIHelper。
func (m *AIHelperManager) GetAIHelper(username string, sessionID string) (*AIHelper, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeExpiredLocked(time.Now())
	userHelpers, exists := m.helpers[username]
	if !exists {
		return nil, false
	}
	entry, exists := userHelpers[sessionID]
	if !exists {
		return nil, false
	}
	entry.lastAccess = time.Now()
	return entry.helper, true
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

func (m *AIHelperManager) removeExpiredLocked(now time.Time) {
	for username, userHelpers := range m.helpers {
		for sessionID, entry := range userHelpers {
			if now.Sub(entry.lastAccess) >= helperIdleTTL {
				delete(userHelpers, sessionID)
			}
		}
		if len(userHelpers) == 0 {
			delete(m.helpers, username)
		}
	}
}

func (m *AIHelperManager) cleanupExpiredHelpers() {
	ticker := time.NewTicker(helperCleanupEvery)
	defer ticker.Stop()
	for now := range ticker.C {
		m.mu.Lock()
		m.removeExpiredLocked(now)
		m.mu.Unlock()
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
		go globalManager.cleanupExpiredHelpers()
	})
	return globalManager
}
