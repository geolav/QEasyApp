package port

import "github.com/geolav/QEasyApp/internal/entity"

type UserRepository interface {
	Create(user entity.User) error
	GetByID(id string) (entity.User, error)
	GetByEmail(email string) (entity.User, error)
}

type UserUseCase interface {
	Register(username, email, password string) (entity.User, error)
	Login(email, password string) (string, error)
	GetProfile(userID string) (entity.User, error)
}
