package entity

type QuizStatus string
type SessionStatus string
type QuestionType string

const (
	QuizStatusDraft     QuizStatus = "draft"
	QuizStatusPublished QuizStatus = "published"
)

const (
	SessionStatusWaiting  SessionStatus = "waiting"
	SessionStatusActive   SessionStatus = "active"
	SessionStatusFinished SessionStatus = "finished"
)

const (
	QuestionTypeSingle   QuestionType = "single_choice"
	QuestionTypeMultiple QuestionType = "multiple_choice"
)
