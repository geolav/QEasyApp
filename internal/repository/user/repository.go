package user

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

func New(db *pgxpool.Pool) port.UserRepository {
	return &repository{db: db}
}

func (r *repository) Create(user entity.User) error {
	_, err := r.db.Exec(context.Background(),
		`INSERT INTO users (id, username, email, password_hash, created_at)
			 VALUES ($1, $2, $3, $4, $5)`, user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt,
	)
	return err
}

func (r *repository) GetByID(id string) (entity.User, error) {
	var user entity.User
	err := r.db.QueryRow(context.Background(),
		`SELECT id, username, email, password_hash, created_at
			FROM users where id=$1`, id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		return entity.User{}, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

func (r *repository) GetByEmail(email string) (entity.User, error) {
	var user entity.User
	err := r.db.QueryRow(context.Background(),
		`SELECT id, username, email, password_hash, created_at
			FROM users where email=$1`, email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		return entity.User{}, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}
