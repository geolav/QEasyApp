package entity

import "time"

type Session struct {
	ID                   string
	QuizID               string
	RoomCode             string
	Status               SessionStatus
	CurrentQuestionIndex int
	StartedAt            *time.Time //  указатель, тк может быть nil до старта
	EndedAt              *time.Time
}
