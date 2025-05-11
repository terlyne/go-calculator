package auth

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("неверные учетные данные")
	ErrUserExists         = errors.New("пользователь уже существует")
)

// Claims представляет данные, хранящиеся в JWT токене
type Claims struct {
	UserID int64  `json:"user_id"`
	Login  string `json:"login"`
	jwt.StandardClaims
}

// Auth предоставляет методы для аутентификации и авторизации
type Auth struct {
	jwtSecret []byte
}

// NewAuth создает новый экземпляр Auth с указанным секретным ключом
func NewAuth(jwtSecret string) *Auth {
	return &Auth{
		jwtSecret: []byte(jwtSecret),
	}
}

// HashPassword хеширует пароль с использованием bcrypt
func (a *Auth) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash проверяет соответствие пароля его хешу
func (a *Auth) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken создает новый JWT токен для пользователя
func (a *Auth) GenerateToken(userID int64, login string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Login:  login,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

// ValidateToken проверяет валидность JWT токена и возвращает данные пользователя
func (a *Auth) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("неверный токен")
	}

	return claims, nil
} 