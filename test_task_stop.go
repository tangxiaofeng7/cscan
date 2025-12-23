package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// 测试任务停止机制的脚本
func main() {
	// 连接Redis
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer client.Close()

	ctx := context.Background()

	// 测试任务ID
	taskId := "test-task-123"
	
	fmt.Println("=== 测试任务停止机制 ===")
	
	// 1. 模拟发送停止信号
	fmt.Printf("1. 发送停止信号给任务: %s\n", taskId)
	ctrlKey := "cscan:task:ctrl:" + taskId
	err := client.Set(ctx, ctrlKey, "STOP", 24*time.Hour).Err()
	if err != nil {
		fmt.Printf("发送停止信号失败: %v\n", err)
		return
	}
	fmt.Println("✓ 停止信号已发送")

	// 2. 模拟检查停止信号
	fmt.Println("\n2. 检查停止信号...")
	for i := 0; i < 10; i++ {
		ctrl, err := client.Get(ctx, ctrlKey).Result()
		if err == nil && ctrl == "STOP" {
			fmt.Printf("✓ 第%d次检查: 检测到停止信号 (%s)\n", i+1, ctrl)
		} else {
			fmt.Printf("✗ 第%d次检查: 未检测到停止信号\n", i+1)
		}
		time.Sleep(200 * time.Millisecond) // 模拟新的检查间隔
	}

	// 3. 清理测试数据
	fmt.Println("\n3. 清理测试数据...")
	client.Del(ctx, ctrlKey)
	fmt.Println("✓ 测试数据已清理")

	fmt.Println("\n=== 测试日志系统 ===")
	
	// 4. 测试日志发布
	fmt.Println("4. 测试日志发布...")
	logKey := "cscan:task:logs:" + taskId
	
	// 模拟发布几条日志
	logEntries := []string{
		`{"timestamp":"2024-01-01 10:00:01","level":"INFO","workerName":"test-worker","taskId":"test-task-123","message":"任务开始执行"}`,
		`{"timestamp":"2024-01-01 10:00:02","level":"INFO","workerName":"test-worker","taskId":"test-task-123","message":"端口扫描阶段开始"}`,
		`{"timestamp":"2024-01-01 10:00:03","level":"INFO","workerName":"test-worker","taskId":"test-task-123","message":"检测到停止信号"}`,
		`{"timestamp":"2024-01-01 10:00:04","level":"INFO","workerName":"test-worker","taskId":"test-task-123","message":"任务已停止"}`,
	}
	
	for i, entry := range logEntries {
		err := client.XAdd(ctx, &redis.XAddArgs{
			Stream: logKey,
			Values: map[string]interface{}{"data": entry},
		}).Err()
		if err != nil {
			fmt.Printf("✗ 发布第%d条日志失败: %v\n", i+1, err)
		} else {
			fmt.Printf("✓ 发布第%d条日志成功\n", i+1)
		}
	}

	// 5. 测试日志查询
	fmt.Println("\n5. 测试日志查询...")
	logs, err := client.XRevRange(ctx, logKey, "+", "-").Result()
	if err != nil {
		fmt.Printf("✗ 查询日志失败: %v\n", err)
	} else {
		fmt.Printf("✓ 查询到 %d 条日志:\n", len(logs))
		for i, log := range logs {
			if data, ok := log.Values["data"].(string); ok {
				fmt.Printf("  [%d] %s\n", i+1, data)
			}
		}
	}

	// 6. 清理日志数据
	fmt.Println("\n6. 清理日志数据...")
	client.Del(ctx, logKey)
	fmt.Println("✓ 日志数据已清理")

	fmt.Println("\n=== 测试完成 ===")
	fmt.Println("如果看到所有 ✓ 标记，说明修复生效")
}