package parser

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestUserID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    uint
		wantErr string
	}{
		{
			name:    "empty user id",
			input:   "",
			wantErr: "user_id 不能为空",
		},
		{
			name:    "invalid user id",
			input:   "abc",
			wantErr: "user_id 格式错误",
		},
		{
			name:    "zero user id",
			input:   "0",
			wantErr: "user_id 格式错误",
		},
		{
			name:  "valid user id",
			input: "123",
			want:  123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UserID(tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestVideoID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    uint
		wantErr string
	}{
		{
			name:    "empty video id",
			input:   "",
			wantErr: "video_id 不能为空",
		},
		{
			name:    "invalid video id",
			input:   "abc",
			wantErr: "video_id 格式错误",
		},
		{
			name:    "zero video id",
			input:   "0",
			wantErr: "video_id 格式错误",
		},
		{
			name:  "valid video id",
			input: "456",
			want:  456,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VideoID(tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestCursor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    uint
		wantErr string
	}{
		{
			name:  "empty cursor",
			input: "",
			want:  0,
		},
		{
			name:    "invalid cursor",
			input:   "abc",
			wantErr: "cursor 格式错误",
		},
		{
			name:  "zero cursor is allowed",
			input: "0",
			want:  0,
		},
		{
			name:  "valid cursor",
			input: "789",
			want:  789,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Cursor(tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestParseHotVideoCursor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    HotVideoCursorValue
		wantErr string
	}{
		{
			name: "empty cursor",
			want: HotVideoCursorValue{},
		},
		{
			name:    "invalid base64 cursor",
			input:   "not-base64",
			wantErr: "cursor 格式错误",
		},
		{
			name:    "invalid json cursor",
			input:   base64.RawURLEncoding.EncodeToString([]byte(`{"id":`)),
			wantErr: "cursor 格式错误",
		},
		{
			name:    "missing id cursor",
			input:   encodeHotVideoCursorForTest(t, HotVideoCursorValue{LikeCount: 10, VisitCount: 20}),
			wantErr: "cursor 格式错误",
		},
		{
			name:  "valid cursor",
			input: encodeHotVideoCursorForTest(t, HotVideoCursorValue{LikeCount: 10, VisitCount: 20, ID: 30}),
			want:  HotVideoCursorValue{LikeCount: 10, VisitCount: 20, ID: 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHotVideoCursor(tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %+v, got %+v", tt.want, got)
			}
		})
	}
}

func TestEncodeHotVideoCursor(t *testing.T) {
	want := HotVideoCursorValue{LikeCount: 11, VisitCount: 22, ID: 33}

	token, err := EncodeHotVideoCursor(want)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	got, err := ParseHotVideoCursor(token)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}

func encodeHotVideoCursorForTest(t *testing.T, cursor HotVideoCursorValue) string {
	t.Helper()

	payload, err := json.Marshal(cursor)
	if err != nil {
		t.Fatalf("failed to marshal cursor: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(payload)
}
