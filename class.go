package common

// DefaultPageSize is used when request PageSize is missing or <= 0.
const DefaultPageSize = 20

// PageInfo is embedded in list requests for pagination (JSON: pageNo, pageSize).
type PageInfo struct {
	PageNum  int `json:"pageNum"`
	PageSize int `json:"pageSize"`
}

// NormalizedPageSize returns PageSize or DefaultPageSize when zero or negative.
func (p PageInfo) NormalizedPageSize() int {
	if p.PageSize < 0 {
		return DefaultPageSize
	}
	return p.PageSize
}
