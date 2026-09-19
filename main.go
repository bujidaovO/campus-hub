package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-sql-driver/mysql"
)

type Activity struct {
	ID       int
	Name     string
	Capacity int
	Open     bool
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "OK",
	})
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	if dbHost == "" {
		log.Fatal("DB_HOST is not set")
	}

	if dbPort == "" {
		log.Fatal("DB_PORT is not set")
	}

	if dbName == "" {
		log.Fatal("DB_NAME is not set")
	}

	if dbUser == "" {
		log.Fatal("DB_USER is not set")
	}

	if dbPassword == "" {
		log.Fatal("DB_PASSWORD is not set")
	}

	cfg := mysql.Config{
		User:   dbUser,
		Passwd: dbPassword,
		Net:    "tcp",
		Addr:   dbHost + ":" + dbPort,
		DBName: dbName,
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

	activities, err := loadActivities(db) //这里的db数据库连接池对象，传入这个函数
	if err != nil {
		log.Fatal("load activities:", err) //Fatal打印之后会直接杀掉程序
	}
	for _, activity := range activities {
		log.Printf("Activity: %+v", activity)
	}

	http.HandleFunc("/health", healthHandler)

	http.ListenAndServe(":8080", nil)
}
