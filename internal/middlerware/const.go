package middlerware

type ctxKey int

const (
	requestIdKey ctxKey = iota
	userIDKey
)
