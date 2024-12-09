package common

type TraceIdstruct struct{}

var TraceIdNameInContext = TraceIdstruct{}

type HttpUrl struct{}

const (
	TraceId = "TID"
	HTTPURL = "url"
)
