package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"cscan/worker"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stat"
)

var (
	serverAddr  = flag.String("s", "localhost:9000", "server address")
	redisAddr   = flag.String("r", "localhost:6379", "redis address for log streaming")
	redisPass   = flag.String("rp", "", "redis password")
	workerName  = flag.String("n", "txf's计算机", "worker name")
	concurrency = flag.Int("c", 5, "concurrency")
)

func main() {
	flag.Parse()

	// 禁用统计日志
	stat.DisableLog()
	logx.DisableStat()

	// 生成Worker名称
	name := *workerName
	if name == "" {
		name = worker.GetWorkerName()
	}

	config := worker.WorkerConfig{
		Name:        name,
		ServerAddr:  *serverAddr,
		RedisAddr:   *redisAddr,
		RedisPass:   *redisPass,
		Concurrency: *concurrency,
		Timeout:     3600,
	}

	w, err := worker.NewWorker(config)
	if err != nil {
		logx.Errorf("create worker failed: %v", err)
		os.Exit(1)
	}

	// 启动Worker
	w.Start()

	fmt.Printf("Worker %s started, connecting to %s\n", name, *serverAddr)
	fmt.Printf("Concurrency: %d\n", *concurrency)

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down worker...")
	w.Stop()
}
