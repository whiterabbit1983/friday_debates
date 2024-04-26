package service

type UserInfo struct {
	Id        uint64
	ChatID    int64
	Name      string
	About     string
	UserID    string
	SessionID string
}

type FlowStep struct {
	Id     uint64
	ChatID int64
	Step   string
}

type SessionInfo struct {
	SessionID string
	ChatID    string
}
