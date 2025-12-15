package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/jwt"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotActive      = errors.New("user account is not active")
)

type AuthUseCase struct {
	userRepo   repository.UserRepository
	jwtManager *jwt.JWTManager
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	jwtManager *jwt.JWTManager,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register регистрирует нового пользователя
func (uc *AuthUseCase) Register(ctx context.Context, email, name, password string) (*entity.User, *jwt.TokenPair, error) {
	// Проверяем существование email
	existingUser, err := uc.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, nil, ErrEmailAlreadyExists
	}

	// Создаем пользователя
	user := &entity.User{
		ID:     uuid.New(),
		Email:  email,
		Name:   name,
		Active: true,
	}

	// Хешируем пароль
	if err := user.HashPassword(password); err != nil {
		return nil, nil, err
	}

	// Сохраняем в БД
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	// Генерируем токены
	tokens, err := uc.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// Login авторизует пользователя
func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*entity.User, *jwt.TokenPair, error) {
	// Получаем пользователя по email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Проверяем пароль
	if !user.CheckPassword(password) {
		return nil, nil, ErrInvalidCredentials
	}

	// Проверяем активность
	if !user.IsActive() {
		return nil, nil, ErrUserNotActive
	}

	// Генерируем токены
	tokens, err := uc.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// RefreshToken обновляет access token
func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	accessToken, _, err := uc.jwtManager.RefreshAccessToken(refreshToken)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// GetCurrentUser получает текущего авторизованного пользователя
func (uc *AuthUseCase) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}
