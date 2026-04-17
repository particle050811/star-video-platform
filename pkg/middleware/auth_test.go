package middleware

import (
	"testing"

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
