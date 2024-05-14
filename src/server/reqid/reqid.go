package reqid

import "context"

func GetRequestId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(0).(string); ok {
		return reqID
	}
	return ""
}
