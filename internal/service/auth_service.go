package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gpslakshan/hireflow/internal/config"
	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
	"github.com/gpslakshan/hireflow/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// Sentinel errors — typed errors the handler can check against
var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidRole        = errors.New("invalid role")
)

type AuthService struct {
	userRepo repository.UserRepo
	cfg      *config.Config
}

func NewAuthService(userRepo repository.UserRepo, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, cfg: cfg}
}

// Register creates a new user account.
func (s *AuthService) Register(req dto.RegisterRequest) (entity.User, error) {
	// 1. Check email is not already taken
	existing, _ := s.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return entity.User{}, ErrEmailTaken
	}

	// 2. Validate role
	role := entity.Role(req.Role)
	if role != entity.RoleCandidate && role != entity.RoleRecruiter {
		return entity.User{}, ErrInvalidRole
	}

	// 3. Hash the password — bcrypt cost 12 is the production standard
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Build the entity and save
	user := entity.User{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         role,
	}

	if err := s.userRepo.Create(&user); err != nil {
		return entity.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login verifies credentials and returns a signed JWT.
func (s *AuthService) Login(req dto.LoginRequest) (string, entity.User, error) {
	// 1. Look up user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil || user == nil {
		return "", entity.User{}, ErrInvalidCredentials
	}

	// 2. Compare submitted password against stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", entity.User{}, ErrInvalidCredentials
	}

	// 3. Mint a JWT
	token, err := s.generateJWT(*user)
	if err != nil {
		return "", entity.User{}, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, *user, nil
}

// generateJWT creates a signed JWT containing user_id and role claims.
func (s *AuthService) generateJWT(user entity.User) (string, error) {
	expiryHours, _ := strconv.Atoi(s.cfg.JWTExpiryHours)

	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    string(user.Role),
		"exp":     time.Now().Add(time.Duration(expiryHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(), // issued at
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) GetByID(id string) (*entity.User, error) {
	return s.userRepo.FindByID(id)
}
