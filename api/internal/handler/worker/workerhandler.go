package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cscan/api/internal/logic"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/pkg/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// WorkerListHandler Worker列表
func WorkerListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewWorkerListLogic(r.Context(), svcCtx)
		resp, err := l.WorkerList()
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// WorkerLogsHandler SSE实时日志推送
func WorkerLogsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置SSE响应头
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		// 发送连接成功消息
		fmt.Fprintf(w, "data: {\"level\":\"INFO\",\"message\":\"日志流连接成功，等待Worker日志...\",\"timestamp\":\"%s\",\"workerName\":\"API\"}\n\n",
			time.Now().Format("2006-01-02 15:04:05"))
		flusher.Flush()

		// 先发送最近的历史日志
		logs, err := svcCtx.RedisClient.XRevRange(r.Context(), "cscan:worker:logs", "+", "-").Result()
		if err == nil && len(logs) > 0 {
			count := 100
			if len(logs) < count {
				count = len(logs)
			}
			for i := count - 1; i >= 0; i-- {
				if data, ok := logs[i].Values["data"].(string); ok {
					fmt.Fprintf(w, "data: %s\n\n", data)
				}
			}
			flusher.Flush()
		}

		// 订阅Redis Pub/Sub
		pubsub := svcCtx.RedisClient.Subscribe(r.Context(), "cscan:worker:logs:realtime")
		defer pubsub.Close()

		ch := pubsub.Channel()

		// 实时推送新日志
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				fmt.Fprintf(w, ": heartbeat\n\n")
				flusher.Flush()
			case msg, ok := <-ch:
				if !ok {
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
				flusher.Flush()
			}
		}
	}
}

// WorkerLogsClearHandler 清空历史日志
func WorkerLogsClearHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := svcCtx.RedisClient.Del(r.Context(), "cscan:worker:logs").Err()
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, &types.BaseResp{Code: 0, Msg: "日志已清空"})
	}
}

// WorkerLogsHistoryHandler 获取历史日志
func WorkerLogsHistoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Limit int `json:"limit"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Limit <= 0 {
			req.Limit = 100
		}

		logs, err := svcCtx.RedisClient.XRevRange(r.Context(), "cscan:worker:logs", "+", "-").Result()
		if err != nil {
			response.Error(w, err)
			return
		}

		result := make([]json.RawMessage, 0)
		count := req.Limit
		if len(logs) < count {
			count = len(logs)
		}
		for i := count - 1; i >= 0; i-- {
			if data, ok := logs[i].Values["data"].(string); ok {
				result = append(result, json.RawMessage(data))
			}
		}

		httpx.OkJson(w, map[string]interface{}{
			"code": 0,
			"list": result,
		})
	}
}
