package port

import "github.com/geolav/QEasyApp/internal/entity"

type QuizRepository interface {
	Create(quiz entity.Quiz) error
	GetByID(id string) (entity.Quiz, error)
	GetByCreatorID(creatorID string) ([]entity.Quiz, error)
	Update(quiz entity.Quiz) error
	Delete(id string) error
	AddQuestion(question entity.Question) error
}

type QuizUseCase interface {
	CreateQuiz(creatorID, title, category string, timePerQuestion int) (entity.Quiz, error)
	AddQuestion(quizID string, question entity.Question) error
	PublishQuiz(quizID, creatorID string) error
	GetQuiz(id string) (entity.Quiz, error)
	GetMyQuizzes(creatorID string) ([]entity.Quiz, error)
}
