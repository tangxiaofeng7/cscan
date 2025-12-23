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

// 日志级别常量
const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

// Redis Key 常量
const (
	WorkerLogStream     = "cscan:worker:logs"
	WorkerLogChannel    = "cscan:worker:logs:realtime"
	TaskLogStreamPrefix = "cscan:task:logs:"
	TaskLogChannelPrefix = "cscan:task:logs:realtime:"
	DefaultMaxLogLen    = 10000
	TaskMaxLogLen       = 5000
)

// LogEntry 日志条目（统一结构）
type LogEntry struct {
	Timestamp  string `json:"timestamp"`
	Level      string `json:"level"`
	WorkerName string `json:"workerName"`
	TaskId     string `json:"taskId,omitempty"`
	Message    string `json:"message"`
}

// Logger 统一日志接口
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// LogPublisher 日志发布器（核心组件）
type LogPublisher struct {
	client     *redis.Client
	workerName string
}

// NewLogPublisher 创建日志发布器
func NewLogPublisher(client *redis.Client, workerName string) *LogPublisher {
	return &LogPublisher{
		client:     client,
		workerName: workerName,
	}
}

// publish 发布日志到 Redis（内部方法）
func (p *LogPublisher) publish(taskId, level, message string) {
	if p.client == nil {
		// Redis连接失败时，至少输出到标准输出，确保日志不丢失
		fmt.Printf("[%s] [%s] [%s] %s: %s\n", 
			time.Now().Format("2006-01-02 15:04:05"), 
			level, 
			p.workerName, 
			taskId, 
			message)
		return
	}

	entry := LogEntry{
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		Level:      level,
		WorkerName: p.workerName,
		TaskId:     taskId,
		Message:    message,
	}

	data, _ := json.Marshal(entry)
	ctx := context.Background()

	// 1. 发布到全局 Pub/Sub（Worker 日志实时推送）
	p.client.Publish(ctx, WorkerLogChannel, string(data))

	// 2. 保存到全局 Stream（Worker 日志历史查询）
	p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: WorkerLogStream,
		MaxLen: DefaultMaxLogLen,
		Approx: true,
		Values: map[string]interface{}{"data": string(data)},
	})

	// 3. 如果有 taskId，同时写入任务专属日志
	if taskId != "" {
		// 任务专属 Stream
		p.client.XAdd(ctx, &redis.XAddArgs{
			Stream: TaskLogStreamPrefix + taskId,
			MaxLen: TaskMaxLogLen,
			Approx: true,
			Values: map[string]interface{}{"data": string(data)},
		})
		// 任务专属 Pub/Sub
		p.client.Publish(ctx, TaskLogChannelPrefix+taskId, string(data))
	}
}

// PublishWorkerLog 发布 Worker 级别日志
func (p *LogPublisher) PublishWorkerLog(level, message string) {
	p.publish("", level, message)
}

// PublishTaskLog 发布任务级别日志
func (p *LogPublisher) PublishTaskLog(taskId, level, message string) {
	p.publish(taskId, level, message)
}

// WorkerLogger Worker 日志记录器
type WorkerLogger struct {
	publisher *LogPublisher
}

// NewWorkerLogger 创建 Worker 日志记录器
func NewWorkerLogger(client *redis.Client, workerName string) *WorkerLogger {
	return &WorkerLogger{
		publisher: NewLogPublisher(client, workerName),
	}
}

// log 内部日志方法，同时输出到控制台和 Redis
// 注意：直接写控制台，不通过 logx，避免被 RedisLogWriter 重复拦截
func (l *WorkerLogger) log(level, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	// 直接输出到控制台（不通过 logx，避免重复）
	fmt.Printf("%s [%s] %s\n", timestamp, level, msg)
	
	// 发布到 Redis
	if l.publisher != nil {
		l.publisher.PublishWorkerLog(level, msg)
	}
}

func (l *WorkerLogger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

func (l *WorkerLogger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

func (l *WorkerLogger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

func (l *WorkerLogger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// TaskLogger 任务日志记录器
type TaskLogger struct {
	publisher *LogPublisher
	taskId    string
}

// NewTaskLogger 创建任务日志记录器
func NewTaskLogger(client *redis.Client, workerName, taskId string) *TaskLogger {
	return &TaskLogger{
		publisher: NewLogPublisher(client, workerName),
		taskId:    taskId,
	}
}

// log 内部日志方法，同时输出到控制台和 Redis
// 注意：直接写控制台，不通过 logx，避免被 RedisLogWriter 重复拦截
func (l *TaskLogger) log(level, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	// 直接输出到控制台（不通过 logx，避免重复）
	fmt.Printf("%s [%s] [Task:%s] %s\n", timestamp, level, l.taskId, msg)
	
	// 发布到 Redis（包含 taskId，会同时写入全局和任务专属日志）
	if l.publisher != nil {
		l.publisher.PublishTaskLog(l.taskId, level, msg)
	}
}

func (l *TaskLogger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

func (l *TaskLogger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

func (l *TaskLogger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

func (l *TaskLogger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// RedisLogWriter 将 logx 日志写入 Redis 的 Writer
// 用于拦截 logx 的输出，同时写入控制台和 Redis
type RedisLogWriter struct {
	publisher *LogPublisher
	stdout    io.Writer
}

// NewRedisLogWriter 创建 Redis 日志写入器
func NewRedisLogWriter(client *redis.Client, workerName string) *RedisLogWriter {
	return &RedisLogWriter{
		publisher: NewLogPublisher(client, workerName),
		stdout:    os.Stdout,
	}
}

// logxEntry logx 的 JSON 日志格式
type logxEntry struct {
	Timestamp string `json:"@timestamp"`
	Level     string `json:"level"`
	Content   string `json:"content"`
}

// Write 实现 io.Writer 接口
func (w *RedisLogWriter) Write(p []byte) (n int, err error) {
	// 先写入控制台
	w.stdout.Write(p)

	if w.publisher == nil || w.publisher.client == nil {
		return len(p), nil
	}

	msg := strings.TrimSpace(string(p))
	if msg == "" {
		return len(p), nil
	}

	var level, message string
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 尝试解析 logx 的 JSON 格式日志
	var logxLog logxEntry
	if err := json.Unmarshal([]byte(msg), &logxLog); err == nil && logxLog.Timestamp != "" {
		// 解析时间
		if t, err := time.Parse("2006-01-02T15:04:05.000-07:00", logxLog.Timestamp); err == nil {
			timestamp = t.Format("2006-01-02 15:04:05")
		} else if t, err := time.Parse("2006-01-02T15:04:05.000Z07:00", logxLog.Timestamp); err == nil {
			timestamp = t.Format("2006-01-02 15:04:05")
		}
		level = strings.ToUpper(logxLog.Level)
		if level == "" {
			level = LevelInfo
		}
		message = logxLog.Content
	} else {
		// 非 JSON 格式，根据内容判断级别
		level = LevelInfo
		if strings.Contains(msg, "[ERROR]") || strings.Contains(strings.ToLower(msg), "error") {
			level = LevelError
		} else if strings.Contains(msg, "[WARN]") || strings.Contains(strings.ToLower(msg), "warn") {
			level = LevelWarn
		} else if strings.Contains(msg, "[DEBUG]") || strings.Contains(strings.ToLower(msg), "debug") {
			level = LevelDebug
		}
		message = msg
	}

	// 构建日志条目并发布
	entry := LogEntry{
		Timestamp:  timestamp,
		Level:      level,
		WorkerName: w.publisher.workerName,
		Message:    message,
	}

	data, _ := json.Marshal(entry)
	ctx := context.Background()

	// 发布到 Pub/Sub
	w.publisher.client.Publish(ctx, WorkerLogChannel, string(data))

	// 保存到 Stream
	w.publisher.client.XAdd(ctx, &redis.XAddArgs{
		Stream: WorkerLogStream,
		MaxLen: DefaultMaxLogLen,
		Approx: true,
		Values: map[string]interface{}{"data": string(data)},
	})

	return len(p), nil
}
