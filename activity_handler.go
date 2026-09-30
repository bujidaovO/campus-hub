package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// healthHandler 是一个 HTTP 处理函数，用于返回服务的健康状态。
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "OK",
	})
}

// /activities
func activitiesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//最小CORS支持
		w.Header().Set("Access-Control-Allow-Origin", "*") //当前测试环境允许其他来源访问

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS") //允许的HTTP方法
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")       //允许客户端发送 Content-Type: application/json
			w.WriteHeader(http.StatusNoContent)
			return //OPTIONS 返回：204 No Content
		}

		//方法选择
		switch r.Method {

		//GET方法
		case http.MethodGet:
			activities, err := loadActivities(db)
			//如果活动读取错误
			if err != nil {
				log.Println("load activities:", err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError) //服务器内部错误（500）

				json.NewEncoder(w).Encode(
					map[string]string{
						"error": "failed to load activities",
					},
				)
				return
			}

			// 查询成功但没有活动时，将 nil 切片转换为空切片，使 JSON 返回 [] 而不是 null。
			if activities == nil {
				activities = []Activity{}
			}

			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(activities); err != nil {
				log.Println("encode activities:", err)
			}

		//POST方法
		case http.MethodPost:
			var req CreateActivityRequest

			//解码并处理错误
			//将请求体按json tag填进结构体
			err := json.NewDecoder(r.Body).Decode(&req) //修改req本身，所以传递地址
			//json解码失败返回400
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)

				json.NewEncoder(w).Encode(
					map[string]string{
						"error": "invalid request body",
					},
				)

				return
			}

			//解码成功，开始校验数据合法性
			//名称校验合法性
			if strings.TrimSpace(req.Name) == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)

				json.NewEncoder(w).Encode(
					map[string]string{
						"error": "name cannot be empty",
					},
				)

				return
			}
			//名额校验合法性
			if req.Capacity <= 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)

				json.NewEncoder(w).Encode(
					map[string]string{
						"error": "capacity must be greater than 0",
					},
				)

				return
			}

			result, err := db.Exec(
				"INSERT INTO activities (name, capacity, open) VALUES (?, ?, ?)",
				req.Name,
				req.Capacity,
				true,
			)
			if err != nil {
				log.Println("insert activity:", err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				json.NewEncoder(w).Encode(
					map[string]string{
						"error": "failed to create activity",
					},
				)

				return
			}

			id, err := result.LastInsertId()
			if err != nil {
				log.Println("get inserted activity id:", err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				json.NewEncoder(w).Encode(
					map[string]string{
						"error": "failed to get created activity id",
					},
				)

				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			json.NewEncoder(w).Encode(
				map[string]int64{
					"id": id,
				},
			)

		//default:MethodNotAllowed
		default:
			w.Header().Set("Allow", "GET, POST, OPTIONS")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)

			json.NewEncoder(w).Encode(
				map[string]string{
					"error": "method not allowed",
				},
			)
		}
	}
}

// /activities/
func activityDetailHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//获得id
		idText := strings.TrimPrefix(r.URL.Path, "/activities/")

		id, err := strconv.Atoi(idText)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(
				map[string]string{
					"error": "invalid activity id",
				},
			)

			return
		}

		activity, err := loadActivity(db, id)

		if err != nil {
			if err == sql.ErrNoRows {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)

				json.NewEncoder(w).Encode(map[string]string{
					"error": "activity not found",
				})
				return
			}

			log.Println("load activity error:", err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]string{
				"error": "failed to load activity",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(activity)
	}
}
