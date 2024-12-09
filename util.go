package common

import "context"

func Recover(ctx context.Context, message string) {
	if r := recover(); r != nil {
		Error(ctx, "panic in go routine", String("callInfo", message))
	}
}
