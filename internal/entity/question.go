package entity

type Question struct {
	ID         string
	QuizID     string
	OrderIndex int
	Type       QuestionType
	Text       string
	ImageURL   string
	TimeLimit  int
	Answers    []Answer
}
