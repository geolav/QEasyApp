package entity

import "time"

type SessionParticipant struct {
	ID          string
	SessionID   string
	UserID      string
	Username    string
	IsOrganizer bool
	TotalScore  int
	JoinedAt    time.Time // TODO *time.Time
}
