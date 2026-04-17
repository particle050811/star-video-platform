package parser

import "testing"

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
