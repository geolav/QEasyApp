package quiz

import (
	"errors"
	"math"
	"math/rand"
	"sort"

	"github.com/geolav/QEasyApp/internal/entity"
	"github.com/google/uuid"
)

func calculateScore(responseTimeMs, timeLimitSec int) int {
	timeLimitMs := timeLimitSec * 1000
	bonus := 1.0 - float64(responseTimeMs)/float64(timeLimitMs)
	bonus = math.Max(0, bonus)
	return int(math.Round(bonus * 1000))
}

func isOrganizer(participans []entity.SessionParticipant, userID string) bool {
	for _, p := range participans {
		if p.UserID == userID && p.IsOrganizer {
			return true
		}
	}
	return false
}

func findQuestion(question []entity.Question, questionID string) (entity.Question, error) {
	for _, q := range question {
		if q.ID == questionID {
			return q, nil
		}
	}
	return entity.Question{}, errors.New("question not found")
}

func findAnswer(answers []entity.Answer, answerID string) (entity.Answer, error) {
	for _, a := range answers {
		if a.ID == answerID {
			return a, nil
		}
	}
	return entity.Answer{}, errors.New("answer not found")
}

func sortByScore(participants []entity.SessionParticipant) {
	sort.Slice(participants, func(i, j int) bool {
		return participants[i].TotalScore > participants[j].TotalScore
	})
}

func generateUUID() string {
	return uuid.New().String()
}

func generateRoomCode() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = letters[rand.Intn(len(letters))]
	}
	return string(code)
}
