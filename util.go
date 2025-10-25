package common

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"runtime"
	"time"
)

func Recover(ctx context.Context, message string, fields ...LogField) {
	if r := recover(); r != nil {
		var buf [1024]byte
		n := runtime.Stack(buf[:], false)
		// Add the existing two LogField to fields
		Error(ctx, message, fields...)
		Error(ctx, fmt.Sprintf("%v\n%s", r, string(buf[:n])))
		// Sleep for 2 seconds when error is found
		time.Sleep(2 * time.Second)
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
