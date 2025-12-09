// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	roleModel "task_Project/model/role"
	userModel "task_Project/model/user"
	"task_Project/task/internal/config"
	"task_Project/task/internal/handler"
	mw "task_Project/task/internal/middleware"
	"task_Project/task/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/task-api.yaml", "the config file")

// wrappers to avoid import cycle in middleware deps typing
type empWrap struct{ e *userModel.Employee }

func (w empWrap) GetEmployeeId() string { return w.e.EmployeeId }
func (w empWrap) GetId() string         { return w.e.Id }

type roleWrap struct{ r *roleModel.Role }

func (w roleWrap) GetPermissions() string { return w.r.Permissions.String }

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 环境变量覆盖配置（用于 Railway / Render / Fly.io 等云平台部署）
	c.ApplyEnvOverrides()

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	// 全局CORS中间件（允许 localhost、127.0.0.1、[::1] 三种前端来源）
	corsMiddleware := mw.NewCorsMiddleware([]string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://[::1]:5173",
	})

	// 全局CORS处理：作为第一个中间件，处理所有请求（包括OPTIONS）

	server.Use(corsMiddleware)

	// 全局JWT中间件：白名单放行，其余统一校验
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			logx.Infof("JWT中间件收到请求: %s %s", r.Method, r.URL.Path)
			// OPTIONS 请求交由 CORS 中间件处理，直接传递给下一个处理器
			if r.Method == http.MethodOptions {
				next(w, r)
				return
			}
			path := r.URL.Path
			// 白名单：登录、注册、登出、静态文件
			if path == "/api/v1/auth/login" || path == "/api/v1/auth/register" || path == "/api/v1/auth/logout" {
				next(w, r)
				return
			}
			// 静态文件请求不需要JWT验证
			if len(path) >= 8 && path[:8] == "/static/" {
				next(w, r)
				return
			}
			ctx.JWTMiddleware.Handle(next)(w, r)
		}
	})

	// 全局权限校验中间件（在路由注册前注入）
	deps := mw.AuthzDeps{
		FindEmployeeByUserID: func(c context.Context, userId string) (interface {
			GetEmployeeId() string
			GetId() string
		}, error) {
			emp, err := ctx.EmployeeModel.FindByUserID(c, userId)
			if err != nil || emp == nil {
				return nil, err
			}
			return empWrap{e: emp}, nil
		},
		ListRolesByEmployeeId: func(c context.Context, employeeId string) ([]interface{ GetPermissions() string }, error) {
			// 改为通过职位查询角色（员工->职位->角色）
			roles, err := ctx.PositionRoleModel.ListRolesByEmployeeId(c, employeeId)
			if err != nil {
				return nil, err
			}
			out := make([]interface{ GetPermissions() string }, 0, len(roles))
			for _, r := range roles {
				out = append(out, roleWrap{r: r})
			}
			return out, nil
		},
	}
	server.Use(mw.NewAuthzMiddleware(deps).Handle)

	// 启动调度器
	go ctx.Scheduler.Start()
	defer ctx.Scheduler.Stop()

	handler.RegisterHandlers(server, ctx)

	// 添加静态文件服务（用于访问上传的文件）
	// 静态文件路由：/static/* -> ./uploads/*
	storageRoot := c.FileStorage.StorageRoot
	if storageRoot == "" {
		storageRoot = "./uploads"
	}
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/static/:path",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			// 使用文件服务器处理静态文件
			http.StripPrefix("/static/", http.FileServer(http.Dir(storageRoot))).ServeHTTP(w, r)
		},
	})
	// 支持多级路径
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/static/:a/:b",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			http.StripPrefix("/static/", http.FileServer(http.Dir(storageRoot))).ServeHTTP(w, r)
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/static/:a/:b/:c",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			http.StripPrefix("/static/", http.FileServer(http.Dir(storageRoot))).ServeHTTP(w, r)
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/static/:a/:b/:c/:d",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			http.StripPrefix("/static/", http.FileServer(http.Dir(storageRoot))).ServeHTTP(w, r)
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/static/:a/:b/:c/:d/:e",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			http.StripPrefix("/static/", http.FileServer(http.Dir(storageRoot))).ServeHTTP(w, r)
		},
	})
	logx.Infof("静态文件服务已启动: /static/* -> %s", storageRoot)

	// 为所有可能的 API 路径添加 OPTIONS 处理（作为后备方案）
	// 注意：这应该不需要，但如果中间件没有拦截，这个可以工作
	corsHandler := func(w http.ResponseWriter, r *http.Request) {
		logx.Infof("OPTIONS路由处理: %s", r.URL.Path)
		corsMiddleware(func(w http.ResponseWriter, r *http.Request) {})(w, r)
	}

	// 通常无需逐个列举；保留通配即可

	// 通配预检，覆盖 /api/v1 下的常见层级
	wildcards := []string{
		"/api/v1/:a",
		"/api/v1/:a/:b",
		"/api/v1/:a/:b/:c",
		"/api/v1/:a/:b/:c/:d",
	}
	for _, p := range wildcards {
		server.AddRoute(rest.Route{Method: http.MethodOptions, Path: p, Handler: corsHandler})
	}

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	logx.Infof("CORS 中间件已注册，允许来源: http://localhost:5173, http://127.0.0.1:5173, http://[::1]:5173")
	server.Start()
}
