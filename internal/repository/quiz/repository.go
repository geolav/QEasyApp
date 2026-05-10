package quiz

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

func New(db *pgxpool.Pool) port.QuizRepository {
	return &repository{db: db}
}

func (r *repository) Create(quiz entity.Quiz) error {
	_, err := r.db.Exec(context.Background(),
		`INSERT INTO quizzes (id, creator_id, title, category, time_per_question, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		quiz.ID, quiz.CreatorID, quiz.Title, quiz.Category, quiz.TimePerQuestion, quiz.Status, quiz.CreatedAt,
	)
	return err
}

func (r *repository) GetByID(id string) (entity.Quiz, error) {
	var quiz entity.Quiz
	err := r.db.QueryRow(context.Background(),
		`SELECT id, creator_id, title, category, time_per_question, status, created_at
			FROM quizzes WHERE id = $1`, id,
	).Scan(&quiz.ID, &quiz.CreatorID, &quiz.Title, &quiz.Category, &quiz.TimePerQuestion, &quiz.Status, &quiz.CreatedAt)

	if err != nil {
		return entity.Quiz{}, fmt.Errorf("quiz not found: %w", err)
	}

	rows, err := r.db.Query(context.Background(),
		`SELECT id, quiz_id, order_index, type, text, image_url, time_limit
			FROM questions WHERE quiz_id = $1 ORDER BY order_index`, id,
	)

	if err != nil {
		return entity.Quiz{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var q entity.Question
		if err := rows.Scan(&q.ID, &q.QuizID, &q.OrderIndex, &q.Type, &q.Text, &q.ImageURL, &q.TimeLimit); err != nil {
			return entity.Quiz{}, err
		}
		answerRows, err := r.db.Query(context.Background(),
			`SELECT id, question_id, text, is_correct
				FROM answers WHERE question_id = $1`, q.ID,
		)
		if err != nil {
			return entity.Quiz{}, err
		}
		for answerRows.Next() {
			var a entity.Answer
			if err := answerRows.Scan(&a.ID, &a.QuestionID, &a.Text, &a.IsCorrect); err != nil {
				answerRows.Close()
				return entity.Quiz{}, err
			}
			q.Answers = append(q.Answers, a)
		}
		answerRows.Close()
		quiz.Questions = append(quiz.Questions, q)
	}
	return quiz, nil
}

func (r *repository) GetByCreatorID(creatorID string) ([]entity.Quiz, error) {
	rows, err := r.db.Query(context.Background(),
		`SELECT id, creator_id, title, category, time_per_question, status, created_at
			FROM quizzes WHERE creator_id = $1 ORDER BY created_at DESC`, creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quizzes []entity.Quiz
	for rows.Next() {
		var quiz entity.Quiz
		if err := rows.Scan(&quiz.ID, &quiz.CreatorID, &quiz.Title, &quiz.Category, &quiz.TimePerQuestion, &quiz.Status, &quiz.CreatedAt); err != nil {
			return nil, err
		}
		quizzes = append(quizzes, quiz)
	}
	return quizzes, nil
}

func (r *repository) Update(quiz entity.Quiz) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE quizzes SET status = $1 WHERE id = $2`, quiz.Status, quiz.ID)
	return err
}

func (r *repository) Delete(id string) error {
	_, err := r.db.Exec(context.Background(),
		`DELETE FROM quizzes WHERE id = $1`, id)
	return err
}

func (r *repository) AddQuestion(question entity.Question) error {
	_, err := r.db.Exec(context.Background(),
		`INSERT INTO questions (id, quiz_id, order_index, type, text, image_url, time_limit)
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		question.ID, question.QuizID, question.OrderIndex,
		question.Type, question.Text, question.ImageURL, question.TimeLimit,
	)
	if err != nil {
		return err
	}

	for _, answer := range question.Answers {
		_, err := r.db.Exec(context.Background(),
			`INSERT INTO answers (id, question_id, text, is_correct)
             VALUES ($1, $2, $3, $4)`,
			answer.ID, answer.QuestionID, answer.Text, answer.IsCorrect)
		if err != nil {
			return err
		}
	}
	return nil
}
