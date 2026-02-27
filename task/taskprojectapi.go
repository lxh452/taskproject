// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"net/http"

	"task_Project/task/internal/config"
	"task_Project/task/internal/handler"
	mw "task_Project/task/internal/middleware"
	"task_Project/task/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/task-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	c.ApplyEnvOverrides()

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	// ========== 注册全局中间件（按执行顺序） ==========

	// 1. CORS（从配置读取白名单，未配置则使用默认开发地址）
	corsOrigins := c.CORS.AllowedOrigins
	if len(corsOrigins) == 0 {
		corsOrigins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	}
	corsMiddleware := mw.NewCorsMiddleware(corsOrigins)
	server.Use(corsMiddleware)

	// 2. 限流
	if ctx.RateLimiter != nil {
		server.Use(ctx.RateLimiter.Handle)
	}

	// 3. 安全响应头
	server.Use(mw.NewSecurityHeadersMiddleware().Handle)

	// 4. CSRF
	if ctx.RedisClient != nil {
		csrfMiddleware := mw.NewCSRFMiddleware(ctx.RedisClient, mw.DefaultCSRFConfig())
		server.Use(csrfMiddleware.Handle)
		logx.Info("[Middleware] CSRF 中间件已注册")
	}

	// 注意：JWT 和 AdminAuth 中间件已通过 .api 文件声明，
	// 由 routes.go 中的 rest.WithMiddlewares 自动应用到对应路由组。

	// ========== 启动后台服务 ==========
	ctx.Scheduler.Start()
	defer ctx.Scheduler.Stop()

	// ========== 注册路由 ==========
	handler.RegisterHandlers(server, ctx)

	// 静态文件服务
	registerStaticRoutes(server, c.FileStorage.StorageRoot)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

// registerStaticRoutes 注册静态文件路由（/static/* -> uploads/）
func registerStaticRoutes(server *rest.Server, storageRoot string) {
	if storageRoot == "" {
		storageRoot = "./uploads"
	}
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(storageRoot)))
	staticHandler := func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}
	// go-zero 不支持通配路由，需要注册多级路径
	for _, p := range []string{
		"/static/:path",
		"/static/:a/:b",
		"/static/:a/:b/:c",
		"/static/:a/:b/:c/:d",
		"/static/:a/:b/:c/:d/:e",
	} {
		server.AddRoute(rest.Route{
			Method:  http.MethodGet,
			Path:    p,
			Handler: staticHandler,
		})
	}
	logx.Infof("[Static] 静态文件服务已启动: /static/* -> %s", storageRoot)
}
