package service

//go:generate go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen

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
