package auth

import (
	"time"

	"database/sql"

	"github.com/golang-jwt/jwt"
	logger "github.com/thalq/gopher_mart/internal/middleware"
	"github.com/thalq/gopher_mart/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db        *sql.DB
	jwtSecret string
}

func NewAuthService(db *sql.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: jwtSecret}
}

func (s *AuthService) GenerateToken(userID int64) string {
	claims := &models.Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return ""
	}

	return tokenString
}

func (s *AuthService) CheckUserExists(username string) (bool, error) {
	var userExists bool
	if err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&userExists); err != nil {
		logger.Sugar.Errorf("Error check user exists: %s", err)
		return false, err
	}
	return userExists, nil
}

func (s *AuthService) Register(login, password string) (int64, error) {
	var userID int64
	hash, err := s.HashPassword(password)
	if err != nil {
		return 0, err
	}
	err = s.db.QueryRow(
		"INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id",
		login, hash,
	).Scan(&userID)
	if err != nil {
		logger.Sugar.Errorf("Error insert user to db: %s", err)
		return 0, err
	}
	return userID, nil
}

func (s *AuthService) CreateUserBalance(userID int64) error {
	_, err := s.db.Exec(`
		INSERT INTO user_balance (user_id, current_balance)
		VALUES ($1, 0.0)
	`, userID)
	if err != nil {
		logger.Sugar.Errorf("Error insert user balance to db: %s", err)
		return err
	}
	return nil
}

func (s *AuthService) Authenticate(login, password string) (bool, int64, error) {
	var storedPassword string
	var userID int64

	err := s.db.QueryRow("SELECT id, password FROM users WHERE username = $1", login).Scan(&userID, &storedPassword)
	if err != nil {
		logger.Sugar.Errorf("Error get user from db: %s", err)
		return false, 0, err
	}
	if !s.CheckPasswordHash(password, storedPassword) {
		return false, 0, nil
	}
	return true, userID, nil
}

func (s *AuthService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
