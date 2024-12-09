package common

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
)

func Recover(ctx context.Context, message string) {
	if r := recover(); r != nil {
		Error(ctx, "panic in go routine", String("callInfo", message))
	}
}
func HmacSha1(keyStr string, message string) string {
	mac := hmac.New(sha1.New, []byte(message))
	mac.Write([]byte([]byte(keyStr)))
	return hex.EncodeToString(mac.Sum(nil))
}
