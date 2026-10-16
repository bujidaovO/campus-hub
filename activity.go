package main

// Pagination 表示活动列表分页信息。
// Page 和 PageSize 表示客户端分页参数，Offset 用于数据库查询。
type Pagination struct {
	Page     int
	PageSize int
	Offset   int
}

// ActivityListResponse 表示 GET /activities 分页后的响应。
type ActivityListResponse struct {
	Data     []Activity `json:"data"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// CreateActivityRequest 表示 POST /activities 的请求体。
type CreateActivityRequest struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}

// UpdateActivityRequest 表示 PATCH /activities/{id} 的请求体
// 使用指针区分“字段未提供”和“字段显式提供零值”的情况。
type UpdateActivityRequest struct {
	Name     *string `json:"name"`
	Capacity *int    `json:"capacity"`
	Open     *bool   `json:"open"`
}

// Activity 结构体用于表示活动信息。
type Activity struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Open     bool   `json:"open"`
}
