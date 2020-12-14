package model

type UserSessions struct {
	Items []Session `json:"items"`
}

type Session struct {
	UserID    string `json:"userId"`
	SessionID string `json:"sessionId"`
	StartTime int64  `json:"startTime"`
	Length    int64  `json:"length"`

	//EndTime int64 // TODO: consider if breaks are allowed or not
	// TODO: other elements like Task, Category etc
}
