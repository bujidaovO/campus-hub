package main

// CreateActivityRequest 表示 POST /activities 的请求体。
type CreateActivityRequest struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}

// Activity 结构体用于表示活动信息。
type Activity struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Open     bool   `json:"open"`
}
