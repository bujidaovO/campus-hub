package main

import (
	"database/sql"
	"fmt"
	"log"
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

func connectDatabase() (*sql.DB, error) {
	dbConfig, err := loadDBConfig()
	if err != nil {
		return nil, err
	}
	//这里创建了一个数据类型为mysql.Config的cfg，用于创建DSN
	cfg := mysql.Config{
		User:   dbConfig.User,
		Passwd: dbConfig.Password,
		Net:    "tcp",
		Addr:   dbConfig.Host + ":" + dbConfig.Port,
		DBName: dbConfig.Name,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	log.Println("database connected") //数据库连接成功
	return db, nil
}
