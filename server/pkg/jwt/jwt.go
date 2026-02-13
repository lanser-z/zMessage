package jwt

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Manager JWT管理器
type Manager struct {
	secret     []byte
	expiration time.Duration
}

// Claims JWT声明
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// NewManager 创建JWT管理器
func NewManager(secret string) *Manager {
	return &Manager{
		secret:     []byte(secret),
		expiration: 7 * 24 * time.Hour, // 7天
	}
}

// GenerateToken 生成Token
func (m *Manager) GenerateToken(userID int64) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   strconv.FormatInt(userID, 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateToken 验证Token
func (m *Manager) ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		return 0, fmt.Errorf("parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 检查是否过期
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			return 0, fmt.Errorf("token expired")
		}
		return claims.UserID, nil
	}

	return 0, fmt.Errorf("invalid token")
}
