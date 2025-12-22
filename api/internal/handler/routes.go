package handler

import (
	"net/http"

	"cscan/api/internal/handler/asset"
	"cscan/api/internal/handler/fingerprint"
	"cscan/api/internal/handler/onlineapi"
	"cscan/api/internal/handler/poc"
	"cscan/api/internal/handler/report"
	"cscan/api/internal/handler/task"
	"cscan/api/internal/handler/user"
	"cscan/api/internal/handler/vul"
	"cscan/api/internal/handler/worker"
	"cscan/api/internal/handler/workspace"
	"cscan/api/internal/middleware"
	"cscan/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	// 公开路由（无需认证）- 只保留登录接口
	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodPost, Path: "/api/v1/login", Handler: user.LoginHandler(svcCtx)},
		},
	)

	// 需要认证的路由
	authMiddleware := middleware.NewAuthMiddleware(svcCtx.Config.Auth.AccessSecret)
	authRoutes := []rest.Route{
		// 用户管理
		{Method: http.MethodPost, Path: "/api/v1/user/list", Handler: user.UserListHandler(svcCtx)},

		// Worker日志（需要认证）
		{Method: http.MethodGet, Path: "/api/v1/worker/logs/stream", Handler: worker.WorkerLogsHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/worker/logs/history", Handler: worker.WorkerLogsHistoryHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/worker/logs/clear", Handler: worker.WorkerLogsClearHandler(svcCtx)},

		// 工作空间
		{Method: http.MethodPost, Path: "/api/v1/workspace/list", Handler: workspace.WorkspaceListHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/workspace/save", Handler: workspace.WorkspaceSaveHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/workspace/delete", Handler: workspace.WorkspaceDeleteHandler(svcCtx)},

		// 资产管理
		{Method: http.MethodPost, Path: "/api/v1/asset/list", Handler: asset.AssetListHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/asset/stat", Handler: asset.AssetStatHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/asset/delete", Handler: asset.AssetDeleteHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/asset/batchDelete", Handler: asset.AssetBatchDeleteHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/asset/history", Handler: asset.AssetHistoryHandler(svcCtx)},

		// 任务管理
		{Method: http.MethodPost, Path: "/api/v1/task/list", Handler: task.MainTaskListHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/create", Handler: task.MainTaskCreateHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/update", Handler: task.MainTaskUpdateHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/delete", Handler: task.MainTaskDeleteHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/batchDelete", Handler: task.MainTaskBatchDeleteHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/retry", Handler: task.MainTaskRetryHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/start", Handler: task.MainTaskStartHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/pause", Handler: task.MainTaskPauseHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/resume", Handler: task.MainTaskResumeHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/stop", Handler: task.MainTaskStopHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/stat", Handler: task.TaskStatHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/profile/list", Handler: task.TaskProfileListHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/profile/save", Handler: task.TaskProfileSaveHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/profile/delete", Handler: task.TaskProfileDeleteHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/task/logs", Handler: task.GetTaskLogsHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/v1/task/logs/stream", Handler: task.TaskLogsStreamHandler(svcCtx)},

		// 漏洞管理
		{Method: http.MethodPost, Path: "/api/v1/vul/list", Handler: vul.VulListHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/vul/detail", Handler: vul.VulDetailHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/vul/stat", Handler: vul.VulStatHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/vul/delete", Handler: vul.VulDeleteHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/vul/batchDelete", Handler: vul.VulBatchDeleteHandler(svcCtx)},

		// Worker管理
		{Method: http.MethodPost, Path: "/api/v1/worker/list", Handler: worker.WorkerListHandler(svcCtx)},

		// 在线API搜索
		{Method: http.MethodPost, Path: "/api/v1/onlineapi/search", Handler: onlineapi.OnlineSearchHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/onlineapi/import", Handler: onlineapi.OnlineImportHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/onlineapi/config/list", Handler: onlineapi.APIConfigListHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/onlineapi/config/save", Handler: onlineapi.APIConfigSaveHandler(svcCtx)},

		// POC标签映射
		{Method: http.MethodPost, Path: "/api/v1/poc/tagmapping/list", Handler: poc.TagMappingListHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/poc/tagmapping/save", Handler: poc.TagMappingSaveHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/poc/tagmapping/delete", Handler: poc.TagMappingDeleteHandler(svcCtx)},

		// 自定义POC
		{Method: http.MethodPost, Path: "/api/v1/poc/custom/list", Handler: poc.CustomPocListHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/poc/custom/save", Handler: poc.CustomPocSaveHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/poc/custom/delete", Handler: poc.CustomPocDeleteHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/poc/custom/batchImport", Handler: poc.CustomPocBatchImportHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/poc/custom/clearAll", Handler: poc.CustomPocClearAllHandler(svcCtx)},

		// Nuclei默认模板
		{Method: http.MethodPost, Path: "/api/v1/poc/nuclei/templates", Handler: poc.NucleiTemplateListHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/poc/nuclei/categories", Handler: poc.NucleiTemplateCategoriesHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/poc/nuclei/sync", Handler: poc.NucleiTemplateSyncHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/poc/nuclei/updateEnabled", Handler: poc.NucleiTemplateUpdateEnabledHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/poc/nuclei/detail", Handler: poc.NucleiTemplateDetailHandler(svcCtx)},

		// 指纹管理
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/list", Handler: fingerprint.FingerprintListHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/save", Handler: fingerprint.FingerprintSaveHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/delete", Handler: fingerprint.FingerprintDeleteHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/categories", Handler: fingerprint.FingerprintCategoriesHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/sync", Handler: fingerprint.FingerprintSyncHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/updateEnabled", Handler: fingerprint.FingerprintUpdateEnabledHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/import", Handler: fingerprint.FingerprintImportHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/importFromFile", Handler: fingerprint.FingerprintImportFromFileHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/clearCustom", Handler: fingerprint.FingerprintClearCustomHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/validate", Handler: fingerprint.FingerprintValidateHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/batchValidate", Handler: fingerprint.FingerprintBatchValidateHandler(svcCtx)},

		// POC验证
		{Method: http.MethodPost, Path: "/api/v1/poc/custom/validate", Handler: poc.PocValidateHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/poc/batchValidate", Handler: poc.PocBatchValidateHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/poc/queryResult", Handler: poc.PocValidationResultQueryHandler(svcCtx)},

		// HTTP服务映射
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/httpservice/list", Handler: fingerprint.HttpServiceMappingListHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/httpservice/save", Handler: fingerprint.HttpServiceMappingSaveHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/fingerprint/httpservice/delete", Handler: fingerprint.HttpServiceMappingDeleteHandler(svcCtx)},

		// 报告管理
		{Method: http.MethodPost, Path: "/api/v1/report/detail", Handler: report.ReportDetailHandler(svcCtx)},
		{Method: http.MethodPost, Path: "/api/v1/report/export", Handler: report.ReportExportHandler(svcCtx)},
	}

	// 为每个路由包装认证中间件
	for i := range authRoutes {
		originalHandler := authRoutes[i].Handler
		authRoutes[i].Handler = func(w http.ResponseWriter, r *http.Request) {
			authMiddleware.Handle(http.HandlerFunc(originalHandler)).ServeHTTP(w, r)
		}
	}

	server.AddRoutes(authRoutes)

	// 需要管理员权限的路由（敏感操作）
	adminRoutes := []rest.Route{
		// 清除日志移到普通认证路由，如需管理员限制可移回此处
	}

	// 为管理员路由包装认证中间件
	for i := range adminRoutes {
		originalHandler := adminRoutes[i].Handler
		adminRoutes[i].Handler = func(w http.ResponseWriter, r *http.Request) {
			authMiddleware.Handle(http.HandlerFunc(originalHandler)).ServeHTTP(w, r)
		}
	}

	if len(adminRoutes) > 0 {
		server.AddRoutes(adminRoutes)
	}

	server.AddRoutes(adminRoutes)
}
