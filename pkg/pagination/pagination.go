package pagination

// Request 分页请求参数
type Request struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"pageSize" json:"pageSize"`
	OrderBy  string `form:"orderBy" json:"orderBy"`
	Sort     string `form:"sort" json:"sort"` // asc / desc
}

// Response 分页响应结构
type Response struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Pages    int         `json:"pages"`
}

// Normalize 标准化分页参数（防越界）
func (r *Request) Normalize() {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.PageSize < 1 || r.PageSize > 100 {
		r.PageSize = 20
	}
	if r.OrderBy == "" {
		r.OrderBy = "id"
	}
	if r.Sort == "" {
		r.Sort = "desc"
	}
}

// Offset 计算 SQL OFFSET
func (r *Request) Offset() int {
	return (r.Page - 1) * r.PageSize
}

// NewResponse 创建分页响应
func NewResponse(list interface{}, total int64, req *Request) *Response {
	pages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		pages++
	}
	return &Response{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Pages:    pages,
	}
}
