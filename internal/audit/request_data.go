package audit

import "context"

type requestDataKey struct{}

// RequestData stores audit metadata which was collected while handling a request
type RequestData struct {
	Operation string
	Metrics   []string
}

// NewRequestData creates an empty RequestData value
func NewRequestData() *RequestData {
	return &RequestData{
		Metrics: make([]string, 0),
	}
}

// WithRequestData returns a context carrying data as audit request metadata
func WithRequestData(ctx context.Context, data *RequestData) context.Context {
	return context.WithValue(ctx, requestDataKey{}, data)
}

// FromContext returns audit request metadata stored in ctx, if any
func FromContext(ctx context.Context) *RequestData {
	data, _ := ctx.Value(requestDataKey{}).(*RequestData)
	return data
}

// AddMetric records name in the RequestData stored in ctx
func AddMetric(ctx context.Context, name string) {
	if name == "" {
		return
	}

	data := FromContext(ctx)
	if data == nil {
		return
	}

	data.Metrics = append(data.Metrics, name)
}

// SetOperation records operation in the RequestData stored in ctx
func SetOperation(ctx context.Context, operation string) {
	if operation == "" {
		return
	}

	data := FromContext(ctx)
	if data == nil {
		return
	}

	data.Operation = operation
}
