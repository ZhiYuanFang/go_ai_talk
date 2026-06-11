package ucg

// PageParams 统一分页入参。
type PageParams struct {
	Page     int
	PageSize int
}

// PageResult 统一分页出参。
type PageResult struct {
	List     interface{} `json:"list"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// CommentsListResult 评论全量列表出参（非分页；超长帖可能 truncated）。
type CommentsListResult struct {
	List      []*CommentDTO `json:"list"`
	Total     int           `json:"total"`
	Truncated bool          `json:"truncated"`
}

const (
	defaultPageSize = 20
	maxPageSize     = 50
)

// NormalizePage 规范化 page/pageSize（page 从 1 起）。
func NormalizePage(page, pageSize int) PageParams {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return PageParams{Page: page, PageSize: pageSize}
}

func pageOffset(p PageParams) int {
	return (p.Page - 1) * p.PageSize
}
