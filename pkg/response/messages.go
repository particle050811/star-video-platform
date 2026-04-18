package response

var codeMsg = mergeCodeMessages(
	commonCodeMsg,
	userCodeMsg,
	relationCodeMsg,
	videoCodeMsg,
	interactionCodeMsg,
	chatCodeMsg,
)

func mergeCodeMessages(groups ...map[int32]string) map[int32]string {
	merged := make(map[int32]string)
	for _, group := range groups {
		for code, msg := range group {
			merged[code] = msg
		}
	}
	return merged
}
