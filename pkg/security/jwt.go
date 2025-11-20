package security

import (
	"errors"
	"time"

	"github.com/afdhali/GolangBlogpostServer/config"
	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService interface {
	GenerateAccessToken(user *entity.User) (string, error)
	GenerateRefreshToken(user *entity.User) (string, error)
	VerifyToken(tokenString string) (jwt.MapClaims, error)
}

type jwtService struct {
	config *config.Config
}

func NewJWTService(cfg *config.Config) JWTService {
	return &jwtService{config: cfg}
}

func (j *jwtService) GenerateAccessToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
        "user_id":  user.ID.String(),
        "email":    user.Email,
        "username": user.Username,
        "role":     string(user.Role),
        "type":     "access",
        "exp":      time.Now().Add(time.Duration(j.config.JWT.AccessTokenExpiry) * time.Second).Unix(),
        "iat":      time.Now().Unix(),
        "jti":      uuid.New().String(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(j.config.JWT.Secret))
}

func (j *jwtService) GenerateRefreshToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
        "user_id": user.ID.String(),
        "type":    "refresh",
        "exp":     time.Now().Add(time.Duration(j.config.JWT.RefreshTokenExpiry) * time.Second).Unix(),
        "iat":     time.Now().Unix(),
        "jti":     uuid.New().String(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(j.config.JWT.Secret))
}

func (j *jwtService) VerifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) ( any, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(j.config.JWT.Secret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}