package common

type TraceIdstruct struct{}

var TraceIdNameInContext = TraceIdstruct{}

type HttpUrl struct{}
type ClientInfo struct{}
type SidStruct struct{}

var SidNameInContext = SidStruct{}

const (
	TraceId    = "TID"
	HTTPURL    = "url"
	CLIENTINFO = "client"
	SID        = "SID"
)
