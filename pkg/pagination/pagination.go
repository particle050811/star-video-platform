package pagination

const (
	DefaultPageNum  int32 = 1
	DefaultPageSize int32 = 20
	MaxPageSize     int32 = 100
)

func Normalize(pageNum, pageSize int32) (offset, limit int) {
	if pageNum <= 0 {
		pageNum = DefaultPageNum
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	limit = int(pageSize)
	offset = int((pageNum - 1) * pageSize)
	return offset, limit
}

func NormalizeLimit(limit int32) int {
	if limit <= 0 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}

	return int(limit)
}
