package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// DBConfig 结构体用于存储数据库连接配置。
type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

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

// loadDBConfig 函数从环境变量中加载数据库配置，并返回一个 DBConfig 结构体和一个 error。
func loadDBConfig() (*DBConfig, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	if host == "" || port == "" || name == "" || user == "" || password == "" {
		return nil, fmt.Errorf("database configuration is not set properly")
	}

	return &DBConfig{
		Host:     host,
		Port:     port,
		Name:     name,
		User:     user,
		Password: password,
	}, nil
}

// loadActivities 函数从数据库中加载活动数据，并返回一个包含所有活动的切片（和error）。
func loadActivities(db *sql.DB) ([]Activity, error) { //db *sql.DB：把 main() 里的数据库连接池对象传进函数。(依旧指针形式)
	rows, err := db.Query(
		"SELECT id, name, capacity, open FROM activities",
	)
	if err != nil {
		return nil, err //查询数据库失败，返回错误
	}
	defer rows.Close()

	var activities []Activity

	for rows.Next() {
		var activity Activity

		err := rows.Scan(
			&activity.ID,
			&activity.Name,
			&activity.Capacity,
			&activity.Open,
		) //Scan() 方法把查询结果的每一行数据映射到 Activity 结构体的字段中,需要注意顺序必须相同
		if err != nil {
			return nil, err
		}

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}

// loadActivity()用于数据库（单个）查询
func loadActivity(db *sql.DB, id int) (Activity, error) {
	var activity Activity

	//
	err := db.QueryRow(`
	SELECT id, name, capacity, open
	FROM activities
	WHERE id = ?
	`, id).Scan(
		&activity.ID,
		&activity.Name,
		&activity.Capacity,
		&activity.Open,
	)

	return activity, err
}

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

func main() {
	dbConfig, err := loadDBConfig()
	if err != nil {
		log.Fatal("load database configuration: ", err)
	}

	cfg := mysql.Config{
		User:   dbConfig.User,
		Passwd: dbConfig.Password,
		Net:    "tcp",
		Addr:   dbConfig.Host + ":" + dbConfig.Port,
		DBName: dbConfig.Name,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN()) //db就是数据库连接池对象
	if err != nil {
		log.Fatal("open database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("ping database:", err)
	}

	log.Println("database connected") //数据库连接成功
	defer db.Close()

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/activities", activitiesHandler(db))
	http.HandleFunc("/activities/", activityDetailHandler(db))

	http.ListenAndServe(":8080", nil)
}
