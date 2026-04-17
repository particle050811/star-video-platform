package pagination

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name           string
		pageNum        int32
		pageSize       int32
		wantOffset int
		wantLimit  int
	}{
		{
			name:       "uses provided values",
			pageNum:    3,
			pageSize:   10,
			wantOffset: 20,
			wantLimit:  10,
		},
		{
			name:       "defaults page num when non positive",
			pageNum:    0,
			pageSize:   10,
			wantOffset: 0,
			wantLimit:  10,
		},
		{
			name:       "defaults page size when non positive",
			pageNum:    2,
			pageSize:   0,
			wantOffset: 20,
			wantLimit:  20,
		},
		{
			name:       "caps page size at max",
			pageNum:    2,
			pageSize:   999,
			wantOffset: 100,
			wantLimit:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOffset, gotLimit := Normalize(tt.pageNum, tt.pageSize)
			if gotOffset != tt.wantOffset {
				t.Fatalf("expected offset %d, got %d", tt.wantOffset, gotOffset)
			}
			if gotLimit != tt.wantLimit {
				t.Fatalf("expected limit %d, got %d", tt.wantLimit, gotLimit)
			}
		})
	}
}
