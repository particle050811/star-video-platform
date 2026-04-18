package middleware

import (
	"context"
	"testing"

	"video-platform/pkg/auth"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func TestBearerToken(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "missing authorization header",
			header: "",
			want:   "",
		},
		{
			name:   "non bearer scheme",
			header: "Basic abc123",
			want:   "",
		},
		{
			name:   "missing token part",
			header: "Bearer",
			want:   "",
		},
		{
			name:   "valid bearer token",
			header: "Bearer token-123",
			want:   "token-123",
		},
		{
			name:   "bearer token with extra spaces",
			header: "  Bearer   token-456   ",
			want:   "token-456",
		},
		{
			name:   "scheme is case insensitive",
			header: "bearer token-789",
			want:   "token-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := app.NewContext(0)
			if tt.header != "" {
				ctx.Request.Header.Set(consts.HeaderAuthorization, tt.header)
			}

			if got := bearerToken(ctx); got != tt.want {
				t.Fatalf("expected token %q, got %q", tt.want, got)
			}
		})
	}
}

func TestJWTAuthWithQueryToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	auth.InitJWT()

	accessToken, _, err := auth.GenerateTokenPair(12)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	ctx := app.NewContext(0)
	ctx.Request.SetRequestURI("/api/v1/chat/ws?access_token=" + accessToken)

	JWTAuthWithQueryToken()(context.Background(), ctx)

	userIDValue, ok := ctx.Get(ContextUserID)
	if !ok {
		t.Fatal("expected user id in context")
	}
	if userIDValue.(uint) != 12 {
		t.Fatalf("expected user id 12, got %v", userIDValue)
	}
}

func TestJWTAuthWithQueryTokenPrefersAuthorizationHeader(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	auth.InitJWT()

	headerToken, _, err := auth.GenerateTokenPair(21)
	if err != nil {
		t.Fatalf("failed to generate header token: %v", err)
	}
	queryToken, _, err := auth.GenerateTokenPair(22)
	if err != nil {
		t.Fatalf("failed to generate query token: %v", err)
	}

	ctx := app.NewContext(0)
	ctx.Request.SetRequestURI("/api/v1/chat/ws?access_token=" + queryToken)
	ctx.Request.Header.Set(consts.HeaderAuthorization, "Bearer "+headerToken)

	JWTAuthWithQueryToken()(context.Background(), ctx)

	userIDValue, ok := ctx.Get(ContextUserID)
	if !ok {
		t.Fatal("expected user id in context")
	}
	if userIDValue.(uint) != 21 {
		t.Fatalf("expected header user id 21, got %v", userIDValue)
	}
}
