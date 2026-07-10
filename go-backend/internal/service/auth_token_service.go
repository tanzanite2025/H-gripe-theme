package service

import (
	"errors"
	"time"

	"tanzanite/internal/domain/auth"
	"tanzanite/internal/domain/user"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TokenUse string `json:"token_use,omitempty"`
	jwt.RegisteredClaims
}

func (s *AuthService) GenerateToken(u *user.User) (string, error) {
	claims := Claims{
		UserID:   u.ID,
		Email:    u.Email,
		Username: u.Username,
		Role:     string(auth.NormalizeRole(u.Role)),
		TokenUse: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtCfg.GetJWTExpireDuration())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

func (s *AuthService) GenerateRefreshToken(u *user.User) (string, error) {
	claims := Claims{
		UserID:   u.ID,
		Email:    u.Email,
		Username: u.Username,
		Role:     string(auth.NormalizeRole(u.Role)),
		TokenUse: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtCfg.GetRefreshExpireDuration())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

func (s *AuthService) AccessTokenMaxAgeSeconds() int {
	return int(s.jwtCfg.GetJWTExpireDuration().Seconds())
}

func (s *AuthService) RefreshTokenMaxAgeSeconds() int {
	return int(s.jwtCfg.GetRefreshExpireDuration().Seconds())
}

// ValidateToken 验证JWT令牌
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtCfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateActiveToken verifies the token and refreshes user state from storage.
func (s *AuthService) ValidateActiveToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenUse != "" && claims.TokenUse != "access" {
		return nil, errors.New("invalid access token")
	}

	currentUser, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	if currentUser.Status != "active" {
		return nil, errors.New("user account is not active")
	}

	claims.Email = currentUser.Email
	claims.Username = currentUser.Username
	claims.Role = string(auth.NormalizeRole(currentUser.Role))

	return claims, nil
}

func (s *AuthService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenUse != "refresh" {
		return nil, errors.New("invalid refresh token")
	}

	currentUser, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	if currentUser.Status != "active" {
		return nil, errors.New("user account is not active")
	}

	claims.Email = currentUser.Email
	claims.Username = currentUser.Username
	claims.Role = string(auth.NormalizeRole(currentUser.Role))

	return claims, nil
}
