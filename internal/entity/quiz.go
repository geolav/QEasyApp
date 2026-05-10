package entity

import "time"

type Quiz struct {
	ID              string
	CreatorID       string
	Title           string
	Category        string
	TimePerQuestion int
	Status          QuizStatus
	Questions       []Question
	CreatedAt       time.Time
}
