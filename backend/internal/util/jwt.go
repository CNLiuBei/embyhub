package util

import (
	"errors"
	"time"

	"embyhub/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT声明
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	RoleID   int    `json:"role_id"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT Token
func GenerateToken(userID int, username string, roleID int) (string, error) {
	cfg := config.GlobalConfig.JWT

	claims := Claims{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(cfg.ExpireHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.GlobalConfig.JWT

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的Token")
}

// RefreshToken 刷新Token
// 如果Token在刷新窗口期内（过期前30分钟），则生成新Token
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查是否在刷新窗口期内（过期前30分钟可刷新）
	refreshWindow := 30 * time.Minute
	if time.Until(claims.ExpiresAt.Time) > refreshWindow {
		// 还没到刷新窗口，返回原Token
		return tokenString, nil
	}

	// 生成新Token
	return GenerateToken(claims.UserID, claims.Username, claims.RoleID)
}

// GetTokenRemainingTime 获取Token剩余有效时间（秒）
func GetTokenRemainingTime(tokenString string) int64 {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return 0
	}
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 0 {
		return 0
	}
	return int64(remaining.Seconds())
}
