package task

import (
	"fmt"
	"net/http"
	"time"

	"cscan/api/internal/logic"
	"cscan/api/internal/middleware"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/pkg/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// MainTaskListHandler 任务列表
func MainTaskListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskListReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskListLogic(r.Context(), svcCtx)
		resp, err := l.MainTaskList(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// MainTaskCreateHandler 创建任务
func MainTaskCreateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskCreateReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskCreateLogic(r.Context(), svcCtx)
		resp, err := l.MainTaskCreate(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// MainTaskDeleteHandler 删除任务
func MainTaskDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskDeleteLogic(r.Context(), svcCtx)
		resp, err := l.MainTaskDelete(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// MainTaskBatchDeleteHandler 批量删除任务
func MainTaskBatchDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskBatchDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskBatchDeleteLogic(r.Context(), svcCtx)
		resp, err := l.MainTaskBatchDelete(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// MainTaskRetryHandler 重试任务
func MainTaskRetryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskRetryReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskRetryLogic(r.Context(), svcCtx)
		resp, err := l.MainTaskRetry(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// MainTaskStartHandler 启动任务
func MainTaskStartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskControlReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskStartLogic(r.Context(), svcCtx)
		resp, err := l.MainTaskStart(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// MainTaskPauseHandler 暂停任务
func MainTaskPauseHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskControlReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskPauseLogic(r.Context(), svcCtx)
		resp, err := l.MainTaskPause(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// MainTaskResumeHandler 继续任务
func MainTaskResumeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskControlReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskResumeLogic(r.Context(), svcCtx)
		resp, err := l.MainTaskResume(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// MainTaskStopHandler 停止任务
func MainTaskStopHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskControlReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskStopLogic(r.Context(), svcCtx)
		resp, err := l.MainTaskStop(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// TaskProfileListHandler 任务配置列表
func TaskProfileListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewTaskProfileListLogic(r.Context(), svcCtx)
		resp, err := l.TaskProfileList()
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// TaskProfileSaveHandler 保存任务配置
func TaskProfileSaveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TaskProfileSaveReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewTaskProfileSaveLogic(r.Context(), svcCtx)
		resp, err := l.TaskProfileSave(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// TaskProfileDeleteHandler 删除任务配置
func TaskProfileDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TaskProfileDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewTaskProfileDeleteLogic(r.Context(), svcCtx)
		resp, err := l.TaskProfileDelete(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// TaskStatHandler 任务统计
func TaskStatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewTaskStatLogic(r.Context(), svcCtx)
		resp, err := l.TaskStat(workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// MainTaskUpdateHandler 更新任务 
func MainTaskUpdateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskUpdateReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskUpdateLogic(r.Context(), svcCtx)
		resp, err := l.MainTaskUpdate(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// GetTaskLogsHandler 获取任务日志 
func GetTaskLogsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetTaskLogsReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewGetTaskLogsLogic(r.Context(), svcCtx)
		resp, err := l.GetTaskLogs(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// TaskLogsStreamHandler SSE实时任务日志推送 
func TaskLogsStreamHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskId := r.URL.Query().Get("taskId")
		if taskId == "" {
			http.Error(w, "taskId is required", http.StatusBadRequest)
			return
		}

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
		fmt.Fprintf(w, "data: {\"level\":\"INFO\",\"message\":\"任务日志流连接成功\",\"timestamp\":\"%s\",\"workerName\":\"API\",\"taskId\":\"%s\"}\n\n",
			time.Now().Format("2006-01-02 15:04:05"), taskId)
		flusher.Flush()

		// 先发送最近的历史日志
		streamKey := "cscan:task:logs:" + taskId
		logs, err := svcCtx.RedisClient.XRevRange(r.Context(), streamKey, "+", "-").Result()
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

		// 订阅任务专属Redis Pub/Sub频道
		pubsubChannel := "cscan:task:logs:realtime:" + taskId
		pubsub := svcCtx.RedisClient.Subscribe(r.Context(), pubsubChannel)
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
