package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

// parsePagination是一个获取和计算分页的函数
// 返回的error只有一种情况：参数不合法（400），error分为三种：page不合法，pageSize不合法，(page-1)与pageSize的积出现了溢出。
func parsePagination(r *http.Request) (Pagination, error) {
	pagination := Pagination{
		Page:     1,
		PageSize: 10,
		Offset:   0,
	}

	query := r.URL.Query()

	//提取并校验page参数
	values, exists := query["page"]
	if exists {
		//如果存在，处理并校验参数：
		// 1、是否有多个参数（400）
		// 2、参数Atoi是否合法（不是小数&&没有字母&&参数不为空&&前后没有空格）（400）
		// 3、参数Atoi之后范围检查（大于等于1）
		if len(values) != 1 {
			return Pagination{}, errors.New("invalid Page parameter")
		}
		page, err := strconv.Atoi(values[0])
		if err != nil {
			return Pagination{}, errors.New("invalid Page parameter")
		}
		//Atoi已经完成，校验page参数范围
		if page < 1 {
			return Pagination{}, errors.New("invalid Page parameter")
		}
		pagination.Page = page
	}

	//提取并校验page_size参数
	values, exists = query["page_size"]
	if exists {
		//如果存在，处理并校验参数：
		// 1、是否有多个参数（400）
		// 2、参数Atoi是否合法（不是小数&&没有字母&&参数不为空&&前后没有空格）（400）
		// 3、参数Atoi之后范围检查（大于等于1小于等于100）
		if len(values) != 1 {
			return Pagination{}, errors.New("invalid PageSize parameter")
		}
		pageSize, err := strconv.Atoi(values[0])
		if err != nil {
			return Pagination{}, errors.New("invalid PageSize parameter")
		}
		//Atoi已经完成，校验page参数范围
		if pageSize < 1 || pageSize > 100 {
			return Pagination{}, errors.New("invalid PageSize parameter")
		}
		pagination.PageSize = pageSize
	}

	if pagination.Page-1 > math.MaxInt/pagination.PageSize {
		return Pagination{}, errors.New("invalid pagination parameters")
	}
	//计算偏移量并返回结果
	pagination.Offset = (pagination.Page - 1) * pagination.PageSize

	return pagination, nil
}

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
			//提取分页参数，计算偏移量保存到Pagination
			pagination, err := parsePagination(r)
			if err != nil {
				writeJSONError(
					w,
					http.StatusBadRequest,
					err.Error(),
				)

				return
			}

			activities, err := loadActivities(
				db,
				pagination.PageSize,
				pagination.Offset,
			)
			if err != nil {
				log.Println("load activities:", err)

				writeJSONError(
					w,
					http.StatusInternalServerError,
					"failed to load activities",
				)

				return
			}

			// 查询成功但没有活动时，将 nil 切片转换为空切片，使 JSON 返回 [] 而不是 null。
			if activities == nil {
				activities = []Activity{}
			}

			//构造标准回应结构体
			response := ActivityListResponse{
				Data:     activities,
				Page:     pagination.Page,
				PageSize: pagination.PageSize,
			}
			//返回当前页面数据并处理可能的编码错误。
			if err := writeJSON(
				w,
				http.StatusOK,
				response,
			); err != nil {
				log.Println("encode response:", err)
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

				//默认分支：未知的内部错误
				log.Println("update activity error:", err)
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
