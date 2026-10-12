package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// healthHandler 是一个 HTTP 处理函数，用于返回服务的健康状态。
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(
		w,
		http.StatusOK,
		map[string]string{
			"status": "OK",
		},
	)
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

				writeJSONError(
					w,
					http.StatusInternalServerError, //服务器内部错误（500）
					"failed to load activities",
				)
				return
			}

			// 查询成功但没有活动时，将 nil 切片转换为空切片，使 JSON 返回 [] 而不是 null。
			if activities == nil {
				activities = []Activity{}
			}

			if err := writeJSON(
				w,
				http.StatusOK,
				activities,
			); err != nil {
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
				writeJSONError(
					w,
					http.StatusBadRequest,
					"invalid request body",
				)

				return
			}

			//解码成功，开始校验数据合法性
			//名称校验合法性
			if strings.TrimSpace(req.Name) == "" {
				writeJSONError(
					w,
					http.StatusBadRequest,
					"name cannot be empty",
				)

				return
			}
			//名额校验合法性
			if req.Capacity <= 0 {
				writeJSONError(
					w,
					http.StatusBadRequest,
					"capacity must be greater than 0",
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

				writeJSONError(
					w,
					http.StatusInternalServerError,
					"failed to create activity",
				)

				return
			}

			id, err := result.LastInsertId()
			if err != nil {
				log.Println("get inserted activity id:", err)

				writeJSONError(
					w,
					http.StatusInternalServerError,
					"failed to get created activity id",
				)

				return
			}

			//成功创建活动，返回成功创建json
			writeJSON(
				w,
				http.StatusCreated,
				map[string]int64{"id": id},
			)

		//default:MethodNotAllowed
		default:
			w.Header().Set("Allow", "GET, POST, OPTIONS")
			writeJSONError(
				w,
				http.StatusMethodNotAllowed,
				"method not allowed",
			)
		}
	}
}

// /activities/
func activityDetailHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		//取得id
		idText := strings.TrimPrefix(r.URL.Path, "/activities/")

		id, err := strconv.Atoi(idText)
		if err != nil {
			writeJSONError(
				w,
				http.StatusBadRequest,
				"invalid activity id",
			)

			return
		}

		//对于某个活动做路径分流处理
		switch r.Method {

		case http.MethodGet:
			//获得详细信息

			activity, err := loadActivity(db, id)

			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					writeJSONError(
						w,
						http.StatusNotFound,
						"activity not found",
					)
					return
				}

				log.Println("load activity error:", err)

				writeJSONError(
					w,
					http.StatusInternalServerError,
					"failed to load activity",
				)
				return
			}

			writeJSON(
				w,
				http.StatusOK,
				activity,
			)

		case http.MethodPatch:
			//修改数据

			//解析请求json为结构体
			var req UpdateActivityRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				writeJSONError(
					w,
					http.StatusBadRequest,
					"invalid request body",
				)

				return
			}

			//校验结构体内数据是否合法：name不允许为空，capacity>0。如果为nil，不做修改。
			if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
				writeJSONError(
					w,
					http.StatusBadRequest,
					"name cannot be empty",
				)

				return
			}

			//校验活动容量
			if req.Capacity != nil && *req.Capacity <= 0 {
				writeJSONError(
					w,
					http.StatusBadRequest,
					"capacity must be greater than 0",
				)

				return
			}

			//校验完成，进入更新数据设计
			err = updateActivity(db, id, req)
			if err != nil {

				if errors.Is(err, sql.ErrNoRows) {
					writeJSONError(
						w,
						http.StatusNotFound,
						"activity not found",
					)
					return
				}

				if errors.Is(err, ErrNoFields) {
					writeJSONError(
						w,
						http.StatusBadRequest,
						"update payload must not be empty",
					)
					return
				}

				writeJSONError(
					w,
					http.StatusInternalServerError,
					"internal server error",
				)
				return
			}
			//查询并返回更新后的数据
			activity, err := loadActivity(db, id)
			if err != nil {
				log.Println("load updated activity error:", err)

				writeJSONError(
					w,
					http.StatusInternalServerError,
					"failed to load updated activity",
				)

				return
			}

			writeJSON(
				w,
				http.StatusOK,
				activity,
			)

		default:
			//说明允许方法，返回405
			w.Header().Set("Allow", "GET, PATCH")
			writeJSONError(
				w,
				http.StatusMethodNotAllowed,
				"method not allowed",
			)
		}
	}
}
