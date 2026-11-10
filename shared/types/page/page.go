package page

// ####################    分页 参数    ########################
type PageRequest struct {
	Page     int `json:"page"       form:"page"          binding:"required"`
	PageSize int `json:"page_size"  form:"page_size"     binding:"required"`
}

type PageResponse struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	PageTotal int   `json:"-"`
	Total     int64 `json:"total"`
}

type ListResponse struct {
	List []map[string]any `json:"list"`
	PageResponse
}
