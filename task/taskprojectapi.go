// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"

	"task_Project/task/internal/config"
	"task_Project/task/internal/handler"
	"task_Project/task/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/taskprojectapi.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	// start background scheduler tasks
	scheduler := svc.NewSchedulerService(ctx)
	go scheduler.StartScheduler()

	// start RabbitMQ task consumers when MQ available
	if ctx.MQ != nil {
		if err := ctx.StartTaskConsumers(); err != nil {
			// log error only; API can still run without MQ
			fmt.Printf("StartTaskConsumers error: %v\n", err)
		}
	}

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
