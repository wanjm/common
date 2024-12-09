package common

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
