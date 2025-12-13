package common

type TraceIdstruct struct{}

var TraceIdNameInContext = TraceIdstruct{}

type HttpUrl struct{}
type ClientInfo struct{}

const (
	TraceId    = "TID"
	HTTPURL    = "url"
	CLIENTINFO = "client"
)
