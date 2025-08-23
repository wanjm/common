package common

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"runtime"
)

func Recover(ctx context.Context, message string) {
	if r := recover(); r != nil {
		fmt.Println("panic in go routine", r)
		var buf [1024]byte
		n := runtime.Stack(buf[0:], false)
		Error(ctx, fmt.Sprint(r), String("callInfo", message), String("stack", string(buf[0:n])))
	}
}
func HmacSha1(keyStr string, message string) string {
	mac := hmac.New(sha1.New, []byte(message))
	mac.Write([]byte([]byte(keyStr)))
	return hex.EncodeToString(mac.Sum(nil))
}

type RpcLogger struct{}

func (rpcrpcLogger *RpcLogger) LogRequest(ctx context.Context, url, request string) {
	Info(ctx, "send rpc request", String("request", request), String("rpcurl", url))
}

func (rpcrpcLogger *RpcLogger) LogResponse(ctx context.Context, url, response string) {
	Info(ctx, "get rpc response", String("response", response), String("rpcurl", url))
}

func (logger *RpcLogger) LogError(ctx context.Context, url, err string) {
	Error(ctx, "rpc error Info", String("err", err), String("rpcurl", url))
}
