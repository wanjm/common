package common

import "testing"

func TestMysqlIdent(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "id", want: "`id`"},
		{in: "order", want: "`order`"},
		{in: "user.name", want: "`user`.`name`"},
		{in: "`key`", want: "`key`"},
		{in: "", want: ""},
	}
	for _, tc := range cases {
		if got := mysqlIdent(tc.in); got != tc.want {
			t.Errorf("mysqlIdent(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
