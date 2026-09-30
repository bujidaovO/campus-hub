package main

import (
	"log"
	"net/http"
)

func main() {
	db, err := connectDatabase()
	if err != nil {
		log.Fatal("connect database: ", err)
	}
	defer db.Close()

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/activities", activitiesHandler(db))
	http.HandleFunc("/activities/", activityDetailHandler(db))

	http.ListenAndServe(":8080", nil)
}
