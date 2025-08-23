package main

import (
	"fmt"

	"github.com/wanjm/common"
)

func a(a common.SqlQueryOptions) {
	b := len(a.QueryFields)
	fmt.Printf("a=%d\n", b)
}
func escape() {
	a(common.SqlQueryOptions{
		QueryFields: []common.Option{
			common.W("id=?", 1),
		},
	})
}
