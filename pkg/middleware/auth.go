package middleware

import (
	"context"
	"errors"
	"strings"

	"video-platform/pkg/auth"
	"video-platform/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

const (
	ContextUserID = "user_id"
)

// JWTAuth 校验 Authorization: Bearer <access_token>，并把用户信息写入请求上下文。
func JWTAuth() app.HandlerFunc {
	return jwtAuthHandler
}

func jwtAuthHandler(ctx context.Context, c *app.RequestContext) {
	tokenString := bearerToken(c)
	if tokenString == "" {
		c.AbortWithStatusJSON(consts.StatusUnauthorized, response.Unauthorized("缺少访问令牌"))
		return
	}

	claims, err := auth.ValidateAccessToken(tokenString)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) {
			c.AbortWithStatusJSON(consts.StatusUnauthorized, response.Error(response.CodeTokenExpired))
			return
		}
		c.AbortWithStatusJSON(consts.StatusUnauthorized, response.Error(response.CodeTokenInvalid))
		return
	}

	c.Set(ContextUserID, claims.UserID)
	c.Next(ctx)
}

func bearerToken(c *app.RequestContext) string {
	authHeader := strings.TrimSpace(string(c.Request.Header.Peek(consts.HeaderAuthorization)))
	if authHeader == "" {
		return ""
	}

	scheme, token, ok := strings.Cut(authHeader, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}

	return strings.TrimSpace(token)
}
