package main

import (
	"flag"
	"fmt"

	"cscan/api/internal/config"
	"cscan/api/internal/handler"
	"cscan/api/internal/svc"
	"cscan/scheduler"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/cscan.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建服务上下文
	svcCtx := svc.NewServiceContext(c)

	// 创建服务组
	group := service.NewServiceGroup()
	defer group.Stop()

	// 创建HTTP服务器
	server := rest.MustNewServer(c.RestConf)
	handler.RegisterHandlers(server, svcCtx)
	group.Add(server)

	// 创建任务调度器服务
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Host,
		Password: c.Redis.Pass,
	})
	schedulerSvc := scheduler.NewSchedulerService(rdb, svcCtx.SyncMethods)
	group.Add(schedulerSvc)

	fmt.Printf("Starting API server at %s:%d...\n", c.Host, c.Port)
	group.Start()
}
