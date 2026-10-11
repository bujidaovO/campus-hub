package main

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
