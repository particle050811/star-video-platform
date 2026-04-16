package auth

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey []byte
var accessTokenExpiry, refreshTokenExpiry time.Duration

func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET 环境变量未设置，请检查 .env 文件")
	}

	secretKey = []byte(secret)
	accessTokenExpiry = 15 * time.Minute    // 访问令牌 15 分钟
	refreshTokenExpiry = 7 * 24 * time.Hour // 刷新令牌 7 天
}

var (
	// ErrTokenExpired Token 已过期
	ErrTokenExpired = errors.New("token 已过期")
	// ErrTokenInvalid Token 无效
	ErrTokenInvalid = errors.New("token 无效")
	// ErrTokenMalformed Token 格式错误
	ErrTokenMalformed = errors.New("token 格式错误")
)

// TokenType Token 类型
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims 自定义 JWT Claims
type Claims struct {
	UserID    uint      `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

func generateToken(userID uint, tokenType TokenType, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// GenerateTokenPair 生成 access_token 和 refresh_token 对
func GenerateTokenPair(userID uint) (accessToken, refreshToken string, err error) {
	accessToken, err = generateToken(userID, AccessToken, accessTokenExpiry)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = generateToken(userID, RefreshToken, refreshTokenExpiry)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ParseToken 解析 Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrTokenInvalid
		}
		return secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// ValidateAccessToken 验证访问令牌
func ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != AccessToken {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// ValidateRefreshToken 验证刷新令牌
func ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != RefreshToken {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// RefreshTokens 使用刷新令牌获取新的令牌对
func RefreshTokens(refreshTokenString string) (newAccessToken, newRefreshToken string, err error) {
	claims, err := ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", err
	}

	return GenerateTokenPair(claims.UserID)
}
