package user

// Entity 表示独立于持久化实现的用户业务对象。
type Entity struct {
	ID       string
	Username string
	Email    string
}
