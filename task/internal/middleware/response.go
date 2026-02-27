package middleware

import (
	"encoding/json"
	"net/http"
)

// errorResponse 统一的 JSON 错误响应格式
type errorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// writeJSONError 写入 JSON 格式的错误响应，保持与 handler 层一致
func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorResponse{Code: code, Msg: msg})
}
