package entity

type Answer struct {
	ID         string
	QuestionID string
	Text       string
	IsCorrect  bool
}
