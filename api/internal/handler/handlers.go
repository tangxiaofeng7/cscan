package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cscan/api/internal/logic"
	"cscan/api/internal/middleware"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"
)

func LoginHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.LoginResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewLoginLogic(r.Context(), ctx)
		resp, _ := l.Login(&req)
		httpResult(w, resp)
	}
}

func UserListHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PageReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.UserListResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewUserListLogic(r.Context(), ctx)
		resp, _ := l.UserList(&req)
		httpResult(w, resp)
	}
}

func WorkspaceListHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PageReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.WorkspaceListResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewWorkspaceListLogic(r.Context(), ctx)
		resp, _ := l.WorkspaceList(&req)
		httpResult(w, resp)
	}
}

func WorkspaceSaveHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WorkspaceSaveReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewWorkspaceSaveLogic(r.Context(), ctx)
		resp, _ := l.WorkspaceSave(&req)
		httpResult(w, resp)
	}
}

func WorkspaceDeleteHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WorkspaceSaveReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewWorkspaceDeleteLogic(r.Context(), ctx)
		resp, _ := l.WorkspaceDelete(&req)
		httpResult(w, resp)
	}
}

func AssetListHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AssetListReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.AssetListResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewAssetListLogic(r.Context(), ctx)
		resp, _ := l.AssetList(&req, workspaceId)
		httpResult(w, resp)
	}
}

func AssetStatHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewAssetStatLogic(r.Context(), ctx)
		resp, _ := l.AssetStat(workspaceId)
		httpResult(w, resp)
	}
}

func AssetDeleteHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AssetDeleteReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewAssetDeleteLogic(r.Context(), ctx)
		resp, _ := l.AssetDelete(&req, workspaceId)
		httpResult(w, resp)
	}
}

func AssetBatchDeleteHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AssetBatchDeleteReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewAssetBatchDeleteLogic(r.Context(), ctx)
		resp, _ := l.AssetBatchDelete(&req, workspaceId)
		httpResult(w, resp)
	}
}

func AssetHistoryHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AssetHistoryReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.AssetHistoryResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewAssetHistoryLogic(r.Context(), ctx)
		resp, _ := l.AssetHistory(&req, workspaceId)
		httpResult(w, resp)
	}
}

func MainTaskListHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskListReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.MainTaskListResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskListLogic(r.Context(), ctx)
		resp, _ := l.MainTaskList(&req, workspaceId)
		httpResult(w, resp)
	}
}

func MainTaskCreateHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskCreateReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskCreateLogic(r.Context(), ctx)
		resp, _ := l.MainTaskCreate(&req, workspaceId)
		httpResult(w, resp)
	}
}

func MainTaskDeleteHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskDeleteReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskDeleteLogic(r.Context(), ctx)
		resp, _ := l.MainTaskDelete(&req, workspaceId)
		httpResult(w, resp)
	}
}

func MainTaskBatchDeleteHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskBatchDeleteReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskBatchDeleteLogic(r.Context(), ctx)
		resp, _ := l.MainTaskBatchDelete(&req, workspaceId)
		httpResult(w, resp)
	}
}

func MainTaskRetryHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskRetryReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskRetryLogic(r.Context(), ctx)
		resp, _ := l.MainTaskRetry(&req, workspaceId)
		httpResult(w, resp)
	}
}

func MainTaskStartHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskControlReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskStartLogic(r.Context(), ctx)
		resp, _ := l.MainTaskStart(&req, workspaceId)
		httpResult(w, resp)
	}
}

func MainTaskPauseHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskControlReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskPauseLogic(r.Context(), ctx)
		resp, _ := l.MainTaskPause(&req, workspaceId)
		httpResult(w, resp)
	}
}

func MainTaskResumeHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskControlReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskResumeLogic(r.Context(), ctx)
		resp, _ := l.MainTaskResume(&req, workspaceId)
		httpResult(w, resp)
	}
}

func MainTaskStopHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MainTaskControlReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewMainTaskStopLogic(r.Context(), ctx)
		resp, _ := l.MainTaskStop(&req, workspaceId)
		httpResult(w, resp)
	}
}

func TaskProfileListHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewTaskProfileListLogic(r.Context(), ctx)
		resp, _ := l.TaskProfileList()
		httpResult(w, resp)
	}
}

func TaskProfileSaveHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TaskProfileSaveReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewTaskProfileSaveLogic(r.Context(), ctx)
		resp, _ := l.TaskProfileSave(&req)
		httpResult(w, resp)
	}
}

func TaskProfileDeleteHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TaskProfileDeleteReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewTaskProfileDeleteLogic(r.Context(), ctx)
		resp, _ := l.TaskProfileDelete(&req)
		httpResult(w, resp)
	}
}

func VulListHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VulListReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.VulListResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewVulListLogic(r.Context(), ctx)
		resp, _ := l.VulList(&req, workspaceId)
		httpResult(w, resp)
	}
}

func VulDeleteHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VulDeleteReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}
		if req.Id == "" {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "ID不能为空"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewVulLogic(r.Context(), ctx)
		resp, _ := l.VulDelete(&req, workspaceId)
		httpResult(w, resp)
	}
}

func VulBatchDeleteHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VulBatchDeleteReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}
		if len(req.Ids) == 0 {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "请选择要删除的漏洞"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewVulLogic(r.Context(), ctx)
		resp, _ := l.VulBatchDelete(&req, workspaceId)
		httpResult(w, resp)
	}
}

func WorkerListHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewWorkerListLogic(r.Context(), ctx)
		resp, _ := l.WorkerList()
		httpResult(w, resp)
	}
}

// ==================== 在线API搜索处理器 ====================

func OnlineSearchHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OnlineSearchReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.OnlineSearchResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewOnlineAPILogic(r.Context(), ctx)
		resp, _ := l.Search(&req, workspaceId)
		httpResult(w, resp)
	}
}

func OnlineImportHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OnlineImportReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewOnlineAPILogic(r.Context(), ctx)
		resp, _ := l.Import(&req, workspaceId)
		httpResult(w, resp)
	}
}

func APIConfigListHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewOnlineAPILogic(r.Context(), ctx)
		resp, _ := l.ConfigList(workspaceId)
		httpResult(w, resp)
	}
}

func APIConfigSaveHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.APIConfigSaveReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewOnlineAPILogic(r.Context(), ctx)
		resp, _ := l.ConfigSave(&req, workspaceId)
		httpResult(w, resp)
	}
}

// WorkerLogsHandler SSE实时日志推送
func WorkerLogsHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置SSE响应头
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Accel-Buffering", "no") // 禁用nginx缓冲

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
		logs, err := ctx.RedisClient.XRevRange(r.Context(), "cscan:worker:logs", "+", "-").Result()
		if err == nil && len(logs) > 0 {
			// 取最近100条，倒序发送
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

		// 订阅Redis Pub/Sub - 在发送历史日志后订阅，避免重复
		pubsub := ctx.RedisClient.Subscribe(r.Context(), "cscan:worker:logs:realtime")
		defer pubsub.Close()

		// 使用ChannelWithSubscriptions确保订阅成功后才开始接收
		ch := pubsub.Channel()

		// 实时推送新日志，使用心跳保持连接
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				// 发送心跳保持连接
				fmt.Fprintf(w, ": heartbeat\n\n")
				flusher.Flush()
			case msg, ok := <-ch:
				if !ok {
					// channel关闭，重新订阅
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
				flusher.Flush()
			}
		}
	}
}

// WorkerLogsClearHandler 清空历史日志
func WorkerLogsClearHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 删除Redis Stream中的所有日志
		err := ctx.RedisClient.Del(r.Context(), "cscan:worker:logs").Err()
		if err != nil {
			httpResult(w, &types.BaseResp{Code: 500, Msg: "清空日志失败: " + err.Error()})
			return
		}
		httpResult(w, &types.BaseResp{Code: 0, Msg: "日志已清空"})
	}
}

// WorkerLogsHistoryHandler 获取历史日志
func WorkerLogsHistoryHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Limit int `json:"limit"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Limit <= 0 {
			req.Limit = 100
		}

		logs, err := ctx.RedisClient.XRevRange(r.Context(), "cscan:worker:logs", "+", "-").Result()
		if err != nil {
			httpResult(w, &types.BaseResp{Code: 500, Msg: "获取日志失败"})
			return
		}

		result := make([]json.RawMessage, 0)
		count := req.Limit
		if len(logs) < count {
			count = len(logs)
		}
		// 倒序遍历，使结果按时间正序排列（旧的在前，新的在后）
		for i := count - 1; i >= 0; i-- {
			if data, ok := logs[i].Values["data"].(string); ok {
				result = append(result, json.RawMessage(data))
			}
		}

		httpResult(w, map[string]interface{}{
			"code": 0,
			"list": result,
		})
	}
}

// ==================== POC标签映射 ====================

func TagMappingListHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewTagMappingListLogic(r.Context(), ctx)
		resp, _ := l.TagMappingList()
		httpResult(w, resp)
	}
}

func TagMappingSaveHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TagMappingSaveReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewTagMappingSaveLogic(r.Context(), ctx)
		resp, _ := l.TagMappingSave(&req)
		httpResult(w, resp)
	}
}

func TagMappingDeleteHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TagMappingDeleteReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewTagMappingDeleteLogic(r.Context(), ctx)
		resp, _ := l.TagMappingDelete(&req)
		httpResult(w, resp)
	}
}

// ==================== 自定义POC ====================

func CustomPocListHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CustomPocListReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.CustomPocListResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewCustomPocListLogic(r.Context(), ctx)
		resp, _ := l.CustomPocList(&req)
		httpResult(w, resp)
	}
}

func CustomPocSaveHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CustomPocSaveReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewCustomPocSaveLogic(r.Context(), ctx)
		resp, _ := l.CustomPocSave(&req)
		httpResult(w, resp)
	}
}

func CustomPocDeleteHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CustomPocDeleteReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewCustomPocDeleteLogic(r.Context(), ctx)
		resp, _ := l.CustomPocDelete(&req)
		httpResult(w, resp)
	}
}

func CustomPocBatchImportHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CustomPocBatchImportReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.CustomPocBatchImportResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewCustomPocBatchImportLogic(r.Context(), ctx)
		resp, _ := l.CustomPocBatchImport(&req)
		httpResult(w, resp)
	}
}

func CustomPocClearAllHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewCustomPocClearAllLogic(r.Context(), ctx)
		resp, _ := l.CustomPocClearAll()
		httpResult(w, resp)
	}
}

// ==================== Nuclei默认模板 ====================

func NucleiTemplateListHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.NucleiTemplateListReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.NucleiTemplateListResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewNucleiTemplateListLogic(r.Context(), ctx)
		resp, _ := l.NucleiTemplateList(&req)
		httpResult(w, resp)
	}
}

func NucleiTemplateCategoriesHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewNucleiTemplateCategoriesLogic(r.Context(), ctx)
		resp, _ := l.NucleiTemplateCategories()
		httpResult(w, resp)
	}
}

func NucleiTemplateSyncHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Force bool `json:"force"` // 强制重新同步（先删除所有模板）
		}
		json.NewDecoder(r.Body).Decode(&req)

		if req.Force {
			// 强制同步：先删除所有模板
			ctx.NucleiTemplateModel.DeleteAll(r.Context())
		}

		go ctx.SyncNucleiTemplates()
		httpResult(w, &types.BaseResp{Code: 0, Msg: "模板同步已开始，请稍后刷新查看"})
	}
}

func NucleiTemplateUpdateEnabledHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.NucleiTemplateUpdateEnabledReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewNucleiTemplateUpdateEnabledLogic(r.Context(), ctx)
		resp, _ := l.UpdateEnabled(&req)
		httpResult(w, resp)
	}
}

func NucleiTemplateDetailHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.NucleiTemplateDetailReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.NucleiTemplateDetailResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewNucleiTemplateDetailLogic(r.Context(), ctx)
		resp, _ := l.GetDetail(&req)
		httpResult(w, resp)
	}
}

// ==================== 指纹管理 ====================

func FingerprintListHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintListReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.FingerprintListResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewFingerprintListLogic(r.Context(), ctx)
		resp, _ := l.FingerprintList(&req)
		httpResult(w, resp)
	}
}

func FingerprintSaveHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintSaveReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewFingerprintSaveLogic(r.Context(), ctx)
		resp, _ := l.FingerprintSave(&req)
		httpResult(w, resp)
	}
}

func FingerprintDeleteHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintDeleteReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewFingerprintDeleteLogic(r.Context(), ctx)
		resp, _ := l.FingerprintDelete(&req)
		httpResult(w, resp)
	}
}

func FingerprintCategoriesHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewFingerprintCategoriesLogic(r.Context(), ctx)
		resp, _ := l.FingerprintCategories()
		httpResult(w, resp)
	}
}

func FingerprintSyncHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintSyncReq
		json.NewDecoder(r.Body).Decode(&req)

		if req.Force {
			// 强制同步：先删除所有内置指纹
			ctx.FingerprintModel.DeleteBuiltin(r.Context())
		}

		go ctx.SyncWappalyzerFingerprints()
		httpResult(w, &types.BaseResp{Code: 0, Msg: "指纹同步已开始，请稍后刷新查看"})
	}
}

func FingerprintUpdateEnabledHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Id      string `json:"id"`
			Enabled bool   `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewFingerprintUpdateEnabledLogic(r.Context(), ctx)
		resp, _ := l.UpdateEnabled(req.Id, req.Enabled)
		httpResult(w, resp)
	}
}

func FingerprintImportHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintImportReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.FingerprintImportResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewFingerprintImportLogic(r.Context(), ctx)
		resp, _ := l.FingerprintImport(&req)
		httpResult(w, resp)
	}
}

func FingerprintImportFromFileHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintImportFromFileReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.FingerprintImportResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewFingerprintImportFromFileLogic(r.Context(), ctx)
		resp, _ := l.FingerprintImportFromFile(&req)
		httpResult(w, resp)
	}
}

func FingerprintClearCustomHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintClearCustomReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.FingerprintClearCustomResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewFingerprintClearCustomLogic(r.Context(), ctx)
		resp, _ := l.FingerprintClearCustom(&req)
		httpResult(w, resp)
	}
}

func FingerprintValidateHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintValidateReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.FingerprintValidateResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewFingerprintValidateLogic(r.Context(), ctx)
		resp, _ := l.FingerprintValidate(&req)
		httpResult(w, resp)
	}
}

func FingerprintBatchValidateHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintBatchValidateReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.FingerprintBatchValidateResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewFingerprintBatchValidateLogic(r.Context(), ctx)
		resp, _ := l.FingerprintBatchValidate(&req)
		httpResult(w, resp)
	}
}

func PocValidateHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PocValidateReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.PocValidateResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewPocValidateLogic(r.Context(), ctx)
		resp, _ := l.PocValidate(&req)
		httpResult(w, resp)
	}
}

func httpResult(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func PocBatchValidateHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PocBatchValidateReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.PocBatchValidateResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewPocBatchValidateLogic(r.Context(), ctx)
		resp, _ := l.PocBatchValidate(&req)
		httpResult(w, resp)
	}
}
func PocValidationResultQueryHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PocValidationResultQueryReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.PocValidationResultQueryResp{Code: 400, Msg: "参数错误"})
			return
		}

		l := logic.NewPocValidationResultQueryLogic(r.Context(), ctx)
		resp, _ := l.PocValidationResultQuery(&req)
		httpResult(w, resp)
	}
}