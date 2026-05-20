package user

import (
	"errors"
	"time"

	"github.com/geolav/QEasyApp/internal/entity"
	"github.com/geolav/QEasyApp/internal/port"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type useCase struct {
	userRepo  port.UserRepository
	jwtSecret string
}

func New(userRepo port.UserRepository, jwtSecret string) port.UserUseCase {
	return &useCase{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (uc *useCase) Register(username, email, password string) (entity.User, error) {
	_, err := uc.userRepo.GetByEmail(email)
	if err == nil {
		return entity.User{}, errors.New("user with this email already exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entity.User{}, err //  TODO мб сделать более норм ошибки через status.Errorf например
	}
	user := entity.User{
		ID:           generateUUID(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	if err := uc.userRepo.Create(user); err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (uc *useCase) Login(email, password string) (string, error) {
	user, err := uc.userRepo.GetByEmail(email)
	if err != nil {
		return "", errors.New("user not found, invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	return token.SignedString([]byte(uc.jwtSecret))
}

func (uc *useCase) GetProfile(userID string) (entity.User, error) {
	return uc.userRepo.GetByID(userID)
}
