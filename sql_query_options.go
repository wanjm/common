package common

type Direction string

const (
	ASCStr  Direction = "ASC"
	DESCStr Direction = "DESC"
)

const (
	ASC  = 1
	DESC = 2
)

type OrderByParam struct {
	Field     string
	Direction Direction
}

type OrderByParams []OrderByParam

type SqlQueryOptions struct {
	QueryFields  []Optioner
	SelectFields []string
	// OmitFields   []string
	Joins       []*MysqlOption
	GroupBy     []string
	OrderFields OrderByParams
	Limit       int
	Offset      int
}

func (o *SqlQueryOptions) SetSkip(skip int) *SqlQueryOptions {
	o.Offset = skip
	return o
}

func (o *SqlQueryOptions) SetLimit(limit int) *SqlQueryOptions {
	o.Limit = limit
	return o
}

func (o *SqlQueryOptions) SetProjection(cols []string) *SqlQueryOptions {
	o.SelectFields = cols
	return o
}
