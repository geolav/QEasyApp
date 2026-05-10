package session

import (
	"context"
	"fmt"

	"github.com/geolav/QEasyApp/internal/entity"
	"github.com/geolav/QEasyApp/internal/port"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) port.SessionRepository {
	return &repository{db: db}
}

func (r *repository) Create(session entity.Session) error {
	_, err := r.db.Exec(context.Background(),
		`INSERT INTO sessions (id, quiz_id, room_code, status, current_question_index)
         VALUES ($1, $2, $3, $4, $5)`,
		session.ID, session.QuizID, session.RoomCode,
		session.Status, session.CurrentQuestionIndex,
	)
	return err
}

func (r *repository) GetByID(id string) (entity.Session, error) {
	var s entity.Session
	err := r.db.QueryRow(context.Background(),
		`SELECT id, quiz_id, room_code, status, current_question_index, started_at, ended_at
         FROM sessions WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.QuizID, &s.RoomCode, &s.Status, &s.CurrentQuestionIndex, &s.StartedAt, &s.EndedAt)
	if err != nil {
		return entity.Session{}, fmt.Errorf("session not found: %w", err)
	}
	return s, nil
}

func (r *repository) GetByRoomCode(roomCode string) (entity.Session, error) {
	var s entity.Session
	err := r.db.QueryRow(context.Background(),
		`SELECT id, quiz_id, room_code, status, current_question_index, started_at, ended_at
         FROM sessions WHERE room_code = $1`,
		roomCode,
	).Scan(&s.ID, &s.QuizID, &s.RoomCode, &s.Status, &s.CurrentQuestionIndex, &s.StartedAt, &s.EndedAt)
	if err != nil {
		return entity.Session{}, fmt.Errorf("session not found: %w", err)
	}
	return s, nil
}

func (r *repository) UpdateStatus(id string, status entity.SessionStatus) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE sessions SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *repository) UpdateCurrentQuestion(id string, index int) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE sessions SET current_question_index = $1 WHERE id = $2`, index, id)
	return err
}

func (r *repository) AddParticipant(participant entity.SessionParticipant) error {
	_, err := r.db.Exec(context.Background(),
		`INSERT INTO session_participants (id, session_id, user_id, is_organizer, total_score, joined_at)
         VALUES ($1, $2, $3, $4, $5, $6)`,
		participant.ID, participant.SessionID, participant.UserID,
		participant.IsOrganizer, participant.TotalScore, participant.JoinedAt,
	)
	return err
}

func (r *repository) GetParticipants(sessionID string) ([]entity.SessionParticipant, error) {
	rows, err := r.db.Query(context.Background(),
		`SELECT sp.id, sp.session_id, sp.user_id, sp.is_organizer, 
                sp.total_score, sp.joined_at, u.username
         FROM session_participants sp
         JOIN users u ON u.id = sp.user_id
         WHERE sp.session_id = $1`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var participants []entity.SessionParticipant
	for rows.Next() {
		var p entity.SessionParticipant
		if err := rows.Scan(&p.ID, &p.SessionID, &p.UserID,
			&p.IsOrganizer, &p.TotalScore, &p.JoinedAt, &p.Username); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	return participants, nil
}

func (r *repository) UpdateParticipantScore(participantID string, score int) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE session_participants SET total_score = total_score + $1 WHERE id = $2`, score, participantID)

	return err
}

func (r *repository) SaveUserAnswer(answer entity.UserAnswer) error {
	_, err := r.db.Exec(context.Background(),
		`INSERT INTO user_answers (id, session_id, participant_id, question_id, answer_id, response_time_ms, score, answered_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		answer.ID, answer.SessionID, answer.ParticipantID, answer.QuestionID,
		answer.AnswerID, answer.ResponseTimeMs, answer.Score, answer.AnsweredAt,
	)
	return err
}

func (r *repository) GetUserAnswers(sessionID, participantID string) ([]entity.UserAnswer, error) {
	rows, err := r.db.Query(context.Background(),
		`SELECT id, session_id, participant_id, question_id, answer_id, response_time_ms, score, answered_at
			FROM user_answers WHERE session_id = $1 AND participant_id = $2`, sessionID, participantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var answers []entity.UserAnswer
	for rows.Next() {
		var a entity.UserAnswer
		if err := rows.Scan(&a.ID, &a.SessionID, &a.ParticipantID, &a.QuestionID, &a.AnswerID, &a.ResponseTimeMs, &a.Score, &a.AnsweredAt); err != nil {
			return nil, err
		}
		answers = append(answers, a)
	}
	return answers, nil
}
