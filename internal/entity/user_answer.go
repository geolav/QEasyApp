package entity

import "time"

type UserAnswer struct {
	ID             string
	SessionID      string
	ParticipantID  string
	QuestionID     string
	AnswerID       string
	ResponseTimeMs int
	Score          int
	AnsweredAt     time.Time // TODO mb *time.Time
}
