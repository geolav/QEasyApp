package quiz

import (
	"errors"
	"time"

	"github.com/geolav/QEasyApp/internal/entity"
	"github.com/geolav/QEasyApp/internal/port"
)

type useCase struct {
	quizRepo port.QuizRepository
}

func New(quizRepo port.QuizRepository) port.QuizUseCase {
	return &useCase{quizRepo: quizRepo}
}

func (uc *useCase) CreateQuiz(creatorID, title, category string, timePerQuestion int) (entity.Quiz, error) {
	quiz := entity.Quiz{
		ID:              generateUUID(),
		CreatorID:       creatorID,
		Title:           title,
		Category:        category,
		TimePerQuestion: timePerQuestion,
		Status:          entity.QuizStatusDraft,
		CreatedAt:       time.Now(),
	}

	if err := uc.quizRepo.Create(quiz); err != nil {
		return entity.Quiz{}, err
	}

	return quiz, nil
}

func (uc *useCase) AddQuestion(quizID string, question entity.Question) error {
	quiz, err := uc.quizRepo.GetByID(quizID)
	if err != nil {
		return errors.New("quiz not found")
	}
	if quiz.Status != entity.QuizStatusDraft {
		return errors.New("cannot edit published quiz")
	}
	question.ID = generateUUID()
	question.QuizID = quizID

	for i := range question.Answers {
		question.Answers[i].ID = generateUUID()
		question.Answers[i].QuestionID = question.ID
	}
	return uc.quizRepo.AddQuestion(question)
}

func (uc *useCase) PublishQuiz(quizID, creatorID string) error {
	quiz, err := uc.quizRepo.GetByID(quizID)
	if err != nil {
		return errors.New("quiz not found")
	}
	if quiz.CreatorID != creatorID {
		return errors.New("not your quiz")
	}
	if len(quiz.Questions) == 0 {
		return errors.New("quiz must have at least one question")
	}

	return uc.quizRepo.Update(entity.Quiz{
		ID:     quizID,
		Status: entity.QuizStatusPublished,
	})
}

func (uc *useCase) GetQuiz(id string) (entity.Quiz, error) {
	return uc.quizRepo.GetByID(id)
}

func (uc *useCase) GetMyQuizzes(creatorID string) ([]entity.Quiz, error) {
	return uc.quizRepo.GetByCreatorID(creatorID)
}
