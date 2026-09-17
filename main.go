package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-sql-driver/mysql"
)

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

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal("open database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("ping database:", err)
	}

	log.Println("database connected")

	defer db.Close()

	http.HandleFunc("/health", healthHandler)

	http.ListenAndServe(":8080", nil)
}
