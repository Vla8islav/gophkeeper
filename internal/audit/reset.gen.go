package audit

import (
	"time"
)

func resetPointer[T any](v *T) {
	if v == nil {
		return
	}

	if resetter, ok := any(v).(interface{ Reset() }); ok {
		resetter.Reset()
		return
	}

	*v = *new(T)
}

func (v *Event) Reset() {
	if v == nil {
		return
	}

	v.Time = *new(UnixTime)
	v.Metrics = v.Metrics[:0]
	v.RemoteAddr = ""
	v.Operation = ""
}

func (v *UnixTime) Reset() {
	if v == nil {
		return
	}

	v.Time = *new(time.Time)
}

func (v *WebSink) Reset() {
	if v == nil {
		return
	}

	resetPointer(v.client)
	v.auditURL = ""
}
