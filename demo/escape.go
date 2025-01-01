package main

import (
	"fmt"

	"gitlab.plaso.cn/module-go/common"
)

func a(a common.SqlQueryOptions) {
	b := len(a.QueryFields)
	fmt.Printf("a=%d\n", b)
}
func escape() {
	a(common.SqlQueryOptions{
		QueryFields: []*common.Where{
			{Query: "id=?", Args: []interface{}{1}},
		},
	})
}
