// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package employee

import (
	"net/http"

	"encoding/json"
	"task_Project/task/internal/logic/employee"
	"task_Project/task/internal/svc"
)

// 获取当前登录用户的员工信息
func GetSelfEmployeeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := employee.NewGetSelfEmployeeLogic(r.Context(), svcCtx)
		resp, err := l.GetSelfEmployee()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = jsonResponse(w, resp)
	}
}

func jsonResponse(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	return enc.Encode(data)
}
