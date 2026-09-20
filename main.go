package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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

// healthHandler 是一个 HTTP 处理函数，用于返回服务的健康状态。
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "OK",
	})
}

// 返回所有信息到前端
func activitiesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//方法检查
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)

			json.NewEncoder(w).Encode(
				map[string]string{
					"error": "method not allowed",
				},
			)

			return
		}

		activities, err := loadActivities(db)
		//活动读取错误。
		if err != nil {
			log.Println("load activities:", err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError) //服务器内部错误（500）

			json.NewEncoder(w).Encode(
				map[string]string{
					"error": "failed to load activities",
				},
			) //为什么这里要使用map?是因为这种键值对的形式适合编写成json吗？
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

	http.ListenAndServe(":8080", nil)
}
