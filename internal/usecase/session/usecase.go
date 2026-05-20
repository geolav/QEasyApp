package session

import (
	"errors"
	"time"

	"github.com/geolav/QEasyApp/internal/entity"
	"github.com/geolav/QEasyApp/internal/port"
)

type useCase struct {
	sessionRepo port.SessionRepository
	quizRepo    port.QuizRepository
}

func New(sessionRepo port.SessionRepository, quizRepo port.QuizRepository) port.SessionUseCase {
	return &useCase{sessionRepo: sessionRepo, quizRepo: quizRepo}
}

func (uc *useCase) CreateSession(quizID, organizerID string) (entity.Session, error) {
	quiz, err := uc.quizRepo.GetByID(quizID)
	if err != nil {
		return entity.Session{}, errors.New("quiz not found")
	}
	if quiz.CreatorID != organizerID {
		return entity.Session{}, errors.New("not your quiz")
	}

	session := entity.Session{
		ID:                   generateUUID(),
		QuizID:               quizID,
		RoomCode:             generateRoomCode(),
		Status:               entity.SessionStatusWaiting,
		CurrentQuestionIndex: 0,
	}

	if err := uc.sessionRepo.Create(session); err != nil {
		return entity.Session{}, err
	}

	organizer := entity.SessionParticipant{
		ID:          generateUUID(),
		SessionID:   session.ID,
		UserID:      organizerID,
		IsOrganizer: true,
		TotalScore:  0,
		JoinedAt:    time.Now(),
	}

	if err := uc.sessionRepo.AddParticipant(organizer); err != nil {
		return entity.Session{}, err
	}

	return session, nil
}

func (uc *useCase) JoinSession(roomCode, userID string) (entity.SessionParticipant, error) {
	session, err := uc.sessionRepo.GetByRoomCode(roomCode)
	if err != nil {
		return entity.SessionParticipant{}, errors.New("session not found")
	}
	if session.Status != entity.SessionStatusWaiting {
		return entity.SessionParticipant{}, errors.New("session is not waiting participant")
	}

	participant := entity.SessionParticipant{
		ID:          generateUUID(),
		SessionID:   session.ID,
		UserID:      userID,
		IsOrganizer: false,
		TotalScore:  0,
		JoinedAt:    time.Now(),
	}

	if err := uc.sessionRepo.AddParticipant(participant); err != nil {
		return entity.SessionParticipant{}, err
	}

	return participant, nil
}

func (uc *useCase) StartSession(sessionID, organizerID string) (entity.Question, error) {
	session, err := uc.sessionRepo.GetByID(sessionID)
	if err != nil {
		return entity.Question{}, errors.New("session not found")
	}

	participants, err := uc.sessionRepo.GetParticipants(sessionID)
	if err != nil {
		return entity.Question{}, err
	}

	if !isOrganizer(participants, organizerID) {
		return entity.Question{}, errors.New("only organizer can start session")
	}

	if err := uc.sessionRepo.UpdateStatus(session.ID, entity.SessionStatusActive); err != nil {
		return entity.Question{}, err
	}

	quiz, err := uc.quizRepo.GetByID(session.QuizID)
	if err != nil {
		return entity.Question{}, err
	}

	if len(quiz.Questions) == 0 {
		return entity.Question{}, errors.New("quiz has no questions")
	}

	return quiz.Questions[0], nil
}

func (uc *useCase) NextQuestion(sessionID, organizerID string) (entity.Question, error) {
	session, err := uc.sessionRepo.GetByID(sessionID)
	if err != nil {
		return entity.Question{}, errors.New("session not found")
	}
	participants, err := uc.sessionRepo.GetParticipants(sessionID)
	if err != nil {
		return entity.Question{}, err
	}
	if !isOrganizer(participants, organizerID) {
		return entity.Question{}, errors.New("not an organizer. only organizer can switch questions")
	}

	quiz, err := uc.quizRepo.GetByID(session.QuizID)
	if err != nil {
		return entity.Question{}, err
	}

	nextIndex := session.CurrentQuestionIndex + 1
	if nextIndex >= len(quiz.Questions) {
		return entity.Question{}, errors.New("no more questions")
	}

	if err := uc.sessionRepo.UpdateCurrentQuestion(sessionID, nextIndex); err != nil {
		return entity.Question{}, err
	}

	return quiz.Questions[nextIndex], nil
}

func (uc *useCase) SubmitAnswer(sessionID, participantID, questionID string, answerIDs []string, responseTimeMs int) (int, error) {
	session, err := uc.sessionRepo.GetByID(sessionID)
	if err != nil {
		return 0, err
	}
	quiz, err := uc.quizRepo.GetByID(session.QuizID)
	if err != nil {
		return 0, err
	}
	question, err := findQuestion(quiz.Questions, questionID)
	if err != nil {
		return 0, err
	}

	timeLimitMs := question.TimeLimit * 1000
	if responseTimeMs > timeLimitMs {
		return 0, errors.New("time is up")
	}

	score := 0
	if question.Type == entity.QuestionTypeSingle {
		// одиночный выбор — как раньше
		if len(answerIDs) > 0 {
			answer, err := findAnswer(question.Answers, answerIDs[0])
			if err == nil && answer.IsCorrect {
				score = calculateScore(responseTimeMs, question.TimeLimit)
			}
		}
	} else {
		// множественный выбор
		// правило: все правильные выбраны И ни одного неправильного
		correctIDs := map[string]bool{}
		for _, a := range question.Answers {
			if a.IsCorrect {
				correctIDs[a.ID] = true
			}
		}

		selectedIDs := map[string]bool{}
		for _, id := range answerIDs {
			selectedIDs[id] = true
		}

		// проверяем что выбраны все правильные
		allCorrectSelected := true
		for id := range correctIDs {
			if !selectedIDs[id] {
				allCorrectSelected = false
				break
			}
		}

		// проверяем что нет лишних неправильных
		noWrongSelected := true
		for id := range selectedIDs {
			if !correctIDs[id] {
				noWrongSelected = false
				break
			}
		}

		if allCorrectSelected && noWrongSelected {
			score = calculateScore(responseTimeMs, question.TimeLimit)
		} else if allCorrectSelected {
			// выбраны все правильные но есть лишние — половина баллов
			score = calculateScore(responseTimeMs, question.TimeLimit) / 2
		}
	}

	// сохраняем первый answer_id для совместимости с БД
	mainAnswerID := ""
	if len(answerIDs) > 0 {
		mainAnswerID = answerIDs[0]
	}

	userAnswer := entity.UserAnswer{
		ID:             generateUUID(),
		SessionID:      sessionID,
		ParticipantID:  participantID,
		QuestionID:     questionID,
		AnswerID:       mainAnswerID,
		AnswerIDs:      answerIDs,
		ResponseTimeMs: responseTimeMs,
		Score:          score,
		AnsweredAt:     time.Now(),
	}

	if err := uc.sessionRepo.SaveUserAnswer(userAnswer); err != nil {
		return 0, err
	}
	if err := uc.sessionRepo.UpdateParticipantScore(participantID, score); err != nil {
		return 0, err
	}

	return score, nil
}

func (uc *useCase) FinishSession(sessionID, organizerID string) error {
	participants, err := uc.sessionRepo.GetParticipants(sessionID)
	if err != nil {
		return err
	}

	if !isOrganizer(participants, organizerID) {
		return errors.New("not an organizer. only organizer can finish session")
	}

	return uc.sessionRepo.UpdateStatus(sessionID, entity.SessionStatusFinished)
}

func (uc *useCase) GetLeaderboard(sessionID string) ([]entity.SessionParticipant, error) {
	participants, err := uc.sessionRepo.GetParticipants(sessionID)
	if err != nil {
		return nil, err
	}

	// фильтруем организатора — он не в лидерборде
	result := make([]entity.SessionParticipant, 0)
	for _, p := range participants {
		if !p.IsOrganizer {
			result = append(result, p)
		}
	}

	// сортируем по убыванию score
	sortByScore(result)

	return result, nil
}
