package port

import "github.com/geolav/QEasyApp/internal/entity"

type SessionRepository interface {
	Create(session entity.Session) error
	GetByID(id string) (entity.Session, error)
	GetByRoomCode(roomCode string) (entity.Session, error)
	UpdateStatus(id string, status entity.SessionStatus) error
	UpdateCurrentQuestion(id string, index int) error

	AddParticipant(participant entity.SessionParticipant) error
	GetParticipants(sessionID string) ([]entity.SessionParticipant, error)
	UpdateParticipantScore(participantID string, score int) error

	SaveUserAnswer(answer entity.UserAnswer) error
	GetUserAnswers(sessionID, participantID string) ([]entity.UserAnswer, error)
}

type SessionUseCase interface {
	CreateSession(quizID, organizerID string) (entity.Session, error)
	JoinSession(roomCode, userID string) (entity.SessionParticipant, error)
	StartSession(sessionID, organizerID string) (entity.Question, error)
	NextQuestion(sessionID, organizerID string) (entity.Question, error)
	//SubmitAnswer возвращает Score
	SubmitAnswer(sessionID, participantID, questionsID, answerID string, responseTimeMs int) (int, error)
	FinishSession(sessionID, participantID string) error
	GetLeaderboard(sessionID string) ([]entity.SessionParticipant, error)
}
