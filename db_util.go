package common

import "go.mongodb.org/mongo-driver/bson"

type Direction string

const (
	ASC  Direction = "ASC"
	DESC Direction = "DESC"
)

type OrderByParam struct {
	Field     string
	Direction Direction
}

type OrderByParams []OrderByParam
type QueryOptions struct {
	SelectFields []string
	OmitFields   []string
	OrderFields  OrderByParams
	Limit        int
	Offset       int
}
type Where struct {
	Query string //缩小范围，仅支持字符串
	Args  []any
}

func W(query string, args ...any) *Where {
	return &Where{Query: query, Args: args}
}

type SqlQueryOptions struct {
	QueryFields  []*Where
	SelectFields []string
	OmitFields   []string
	OrderFields  OrderByParams
	Limit        int
	Offset       int
}

type MongoQueryOptions struct {
	QueryFields  bson.M
	SelectFields bson.M
	OrderFields  bson.M
	Limit        int
	Offset       int
}
