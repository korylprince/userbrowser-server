package httpapi

const (
	headerContentType = "Content-Type"
	mediaTypeJSON     = "application/json"
)

type contextKey int

const (
	contextKeyUser contextKey = iota
	contextKeyLogData
)
