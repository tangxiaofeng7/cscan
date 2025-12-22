package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// LogEntry 日志条目
type LogEntry struct {
	Timestamp  string `json:"timestamp"`
	Level      string `json:"level"`
	WorkerName string `json:"workerName"`
	TaskId     string `json:"taskId,omitempty"` // 新增：任务ID，用于关联日志与任务
	Message    string `json:"message"`
}

// RedisLogWriter 将日志写入Redis的Writer，同时输出到控制台
type RedisLogWriter struct {
	client     *redis.Client
	workerName string
	streamKey  string
	maxLen     int64
	stdout     io.Writer
}

// NewRedisLogWriter 创建Redis日志写入器
func NewRedisLogWriter(client *redis.Client, workerName string) *RedisLogWriter {
	return &RedisLogWriter{
		client:     client,
		workerName: workerName,
		streamKey:  "cscan:worker:logs",
		maxLen:     10000, // 最多保留10000条日志
		stdout:     os.Stdout,
	}
}

// logxEntry logx的JSON日志格式
type logxEntry struct {
	Timestamp string `json:"@timestamp"`
	Level     string `json:"level"`
	Content   string `json:"content"`
}

// Write 实现io.Writer接口，同时写入控制台和Redis
func (w *RedisLogWriter) Write(p []byte) (n int, err error) {
	// 先写入控制台
	w.stdout.Write(p)

	if w.client == nil {
		return len(p), nil
	}

	msg := strings.TrimSpace(string(p))
	if msg == "" {
		return len(p), nil
	}

	var entry LogEntry
	entry.WorkerName = w.workerName

	// 尝试解析logx的JSON格式日志
	var logxLog logxEntry
	if err := json.Unmarshal([]byte(msg), &logxLog); err == nil && logxLog.Timestamp != "" {
		// 成功解析logx JSON格式，提取时间和级别
		// 解析时间 "2025-12-17T16:34:54.670+08:00" -> "2025-12-17 16:34:54"
		if t, err := time.Parse("2006-01-02T15:04:05.000-07:00", logxLog.Timestamp); err == nil {
			entry.Timestamp = t.Format("2006-01-02 15:04:05")
		} else if t, err := time.Parse("2006-01-02T15:04:05.000Z07:00", logxLog.Timestamp); err == nil {
			entry.Timestamp = t.Format("2006-01-02 15:04:05")
		} else {
			entry.Timestamp = time.Now().Format("2006-01-02 15:04:05")
		}
		entry.Level = strings.ToUpper(logxLog.Level)
		if entry.Level == "" {
			entry.Level = "INFO"
		}
		entry.Message = logxLog.Content
	} else {
		// 非JSON格式，使用原有逻辑
		entry.Timestamp = time.Now().Format("2006-01-02 15:04:05")
		entry.Level = "INFO"
		if strings.Contains(msg, "[ERROR]") || strings.Contains(strings.ToLower(msg), "error") {
			entry.Level = "ERROR"
		} else if strings.Contains(msg, "[WARN]") || strings.Contains(strings.ToLower(msg), "warn") {
			entry.Level = "WARN"
		} else if strings.Contains(msg, "[DEBUG]") || strings.Contains(strings.ToLower(msg), "debug") {
			entry.Level = "DEBUG"
		}
		entry.Message = msg
	}

	data, _ := json.Marshal(entry)

	ctx := context.Background()
	
	// 发布到Pub/Sub频道（用于实时推送）
	if err := w.client.Publish(ctx, "cscan:worker:logs:realtime", string(data)).Err(); err != nil {
		fmt.Fprintf(w.stdout, "[LogWriter] Publish to Redis failed: %v\n", err)
	}

	// 同时保存到Stream（用于历史查询）
	if err := w.client.XAdd(ctx, &redis.XAddArgs{
		Stream: w.streamKey,
		MaxLen: w.maxLen,
		Approx: true,
		Values: map[string]interface{}{
			"data": string(data),
		},
	}).Err(); err != nil {
		fmt.Fprintf(w.stdout, "[LogWriter] XAdd to Redis failed: %v\n", err)
	}

	return len(p), nil
}

// PublishLog 发布日志到Redis
func PublishLog(client *redis.Client, workerName, level, message string) {
	if client == nil {
		return
	}

	entry := LogEntry{
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		Level:      level,
		WorkerName: workerName,
		Message:    message,
	}

	data, _ := json.Marshal(entry)

	ctx := context.Background()
	// 发布到Pub/Sub频道（用于实时推送）
	client.Publish(ctx, "cscan:worker:logs:realtime", string(data))

	// 同时保存到Stream（用于历史查询）
	client.XAdd(ctx, &redis.XAddArgs{
		Stream: "cscan:worker:logs",
		MaxLen: 10000,
		Approx: true,
		Values: map[string]interface{}{
			"data": string(data),
		},
	})
}

// PublishTaskLog 发布任务日志到Redis
// 同时写入全局Stream和任务专属Stream
// 同时发布到全局Pub/Sub和任务专属Pub/Sub
func PublishTaskLog(client *redis.Client, workerName, taskId, level, message string) {
	if client == nil {
		return
	}

	entry := LogEntry{
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		Level:      level,
		WorkerName: workerName,
		TaskId:     taskId,
		Message:    message,
	}

	data, _ := json.Marshal(entry)
	ctx := context.Background()

	// 1. 发布到全局Pub/Sub频道（用于Worker页面实时推送）
	client.Publish(ctx, "cscan:worker:logs:realtime", string(data))

	// 2. 保存到全局Stream（用于Worker页面历史查询）
	client.XAdd(ctx, &redis.XAddArgs{
		Stream: "cscan:worker:logs",
		MaxLen: 10000,
		Approx: true,
		Values: map[string]interface{}{
			"data": string(data),
		},
	})

	// 3. 保存到任务专属Stream（用于任务日志查询）
	if taskId != "" {
		client.XAdd(ctx, &redis.XAddArgs{
			Stream: "cscan:task:logs:" + taskId,
			MaxLen: 5000, // 每个任务最多保留5000条日志
			Approx: true,
			Values: map[string]interface{}{
				"data": string(data),
			},
		})

		// 4. 发布到任务专属Pub/Sub频道（用于任务日志实时推送）
		client.Publish(ctx, "cscan:task:logs:realtime:"+taskId, string(data))
	}
}

// WorkerLogger Worker日志记录器
type WorkerLogger struct {
	client     *redis.Client
	workerName string
}

// NewWorkerLogger 创建Worker日志记录器
func NewWorkerLogger(client *redis.Client, workerName string) *WorkerLogger {
	return &WorkerLogger{
		client:     client,
		workerName: workerName,
	}
}

func (l *WorkerLogger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	PublishLog(l.client, l.workerName, "INFO", msg)
}

func (l *WorkerLogger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	PublishLog(l.client, l.workerName, "ERROR", msg)
}

func (l *WorkerLogger) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	PublishLog(l.client, l.workerName, "WARN", msg)
}

func (l *WorkerLogger) Debug(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	PublishLog(l.client, l.workerName, "DEBUG", msg)
}

// TaskLogger 任务日志记录器，支持任务级别日志
type TaskLogger struct {
	client     *redis.Client
	workerName string
	taskId     string
}

// NewTaskLogger 创建任务日志记录器
func NewTaskLogger(client *redis.Client, workerName, taskId string) *TaskLogger {
	return &TaskLogger{
		client:     client,
		workerName: workerName,
		taskId:     taskId,
	}
}

func (l *TaskLogger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	PublishTaskLog(l.client, l.workerName, l.taskId, "INFO", msg)
}

func (l *TaskLogger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	PublishTaskLog(l.client, l.workerName, l.taskId, "ERROR", msg)
}

func (l *TaskLogger) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	PublishTaskLog(l.client, l.workerName, l.taskId, "WARN", msg)
}

func (l *TaskLogger) Debug(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	PublishTaskLog(l.client, l.workerName, l.taskId, "DEBUG", msg)
}
