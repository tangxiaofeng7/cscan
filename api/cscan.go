package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"cscan/api/internal/config"
	"cscan/api/internal/handler"
	"cscan/api/internal/svc"
	"cscan/scheduler"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/cscan-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建服务上下文
	ctx := svc.NewServiceContext(c)

	// 创建HTTP服务器
	server := rest.MustNewServer(c.RestConf)
	handler.RegisterHandlers(server, ctx)

	// 创建任务调度器
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Host,
		Password: c.Redis.Pass,
	})
	sched := scheduler.NewScheduler(rdb)
	sched.Start()

	// 创建定时任务管理器
	cronManager := scheduler.NewCronManager(sched, rdb)
	cronManager.LoadTasks(context.Background())

	// 启动HTTP服务器
	go func() {
		fmt.Printf("Starting API server at %s:%d...\n", c.Host, c.Port)
		server.Start()
	}()

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")
	server.Stop()
	sched.Stop()
}
