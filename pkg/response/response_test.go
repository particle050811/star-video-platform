package response

import "testing"

func TestSuccess(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		resp := Success()
		if resp.Code != CodeSuccess {
			t.Fatalf("expected code %d, got %d", CodeSuccess, resp.Code)
		}
		if resp.Msg != "成功" {
			t.Fatalf("expected message %q, got %q", "成功", resp.Msg)
		}
	})

	t.Run("custom message", func(t *testing.T) {
		resp := Success("操作成功")
		if resp.Code != CodeSuccess {
			t.Fatalf("expected code %d, got %d", CodeSuccess, resp.Code)
		}
		if resp.Msg != "操作成功" {
			t.Fatalf("expected message %q, got %q", "操作成功", resp.Msg)
		}
	})
}

func TestError(t *testing.T) {
	t.Run("mapped message", func(t *testing.T) {
		resp := Error(CodeUserNotFound)
		if resp.Code != CodeUserNotFound {
			t.Fatalf("expected code %d, got %d", CodeUserNotFound, resp.Code)
		}
		if resp.Msg != "用户不存在" {
			t.Fatalf("expected message %q, got %q", "用户不存在", resp.Msg)
		}
	})

	t.Run("custom message overrides mapping", func(t *testing.T) {
		resp := Error(CodeUserNotFound, "自定义错误")
		if resp.Code != CodeUserNotFound {
			t.Fatalf("expected code %d, got %d", CodeUserNotFound, resp.Code)
		}
		if resp.Msg != "自定义错误" {
			t.Fatalf("expected message %q, got %q", "自定义错误", resp.Msg)
		}
	})

	t.Run("unknown code falls back to default message", func(t *testing.T) {
		resp := Error(9999)
		if resp.Code != 9999 {
			t.Fatalf("expected code %d, got %d", 9999, resp.Code)
		}
		if resp.Msg != "未知错误" {
			t.Fatalf("expected message %q, got %q", "未知错误", resp.Msg)
		}
	})
}

func TestResponseHelpers(t *testing.T) {
	tests := []struct {
		name     string
		resp     func() int32
		wantCode int32
	}{
		{
			name: "param error",
			resp: func() int32 { return ParamError().Code },
			wantCode: CodeParamError,
		},
		{
			name: "unauthorized",
			resp: func() int32 { return Unauthorized().Code },
			wantCode: CodeUnauthorized,
		},
		{
			name: "forbidden",
			resp: func() int32 { return Forbidden().Code },
			wantCode: CodeForbidden,
		},
		{
			name: "not found",
			resp: func() int32 { return NotFound().Code },
			wantCode: CodeNotFound,
		},
		{
			name: "internal error",
			resp: func() int32 { return InternalError().Code },
			wantCode: CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp(); got != tt.wantCode {
				t.Fatalf("expected code %d, got %d", tt.wantCode, got)
			}
		})
	}
}
