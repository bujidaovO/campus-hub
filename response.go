package main

import (
	"encoding/json"
	"net/http"
)

// writeJSON 输出指定状态码的 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

// writeJSONError 输出统一格式的 JSON 错误响应。
func writeJSONError(w http.ResponseWriter, status int, message string) error {
	return writeJSON(
		w,
		status,
		map[string]string{
			"error": message,
		},
	)
}
