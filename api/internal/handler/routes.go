package handler

import (
	"cscan/api/internal/middleware"
	"cscan/api/internal/svc"
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, ctx *svc.ServiceContext) {
	// 公开路由（无需认证）
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  "POST",
				Path:    "/api/v1/login",
				Handler: LoginHandler(ctx),
			},
			{
				Method:  "GET",
				Path:    "/api/v1/worker/logs/stream",
				Handler: WorkerLogsHandler(ctx),
			},
			{
				Method:  "POST",
				Path:    "/api/v1/worker/logs/history",
				Handler: WorkerLogsHistoryHandler(ctx),
			},
			{
				Method:  "POST",
				Path:    "/api/v1/worker/logs/clear",
				Handler: WorkerLogsClearHandler(ctx),
			},
		},
	)

	// 需要认证的路由 - 使用中间件包装
	authMiddleware := middleware.NewAuthMiddleware(ctx.Config.Auth.AccessSecret)
	authRoutes := []rest.Route{
		// 用户管理
		{Method: "POST", Path: "/api/v1/user/list", Handler: UserListHandler(ctx)},
		// 工作空间
		{Method: "POST", Path: "/api/v1/workspace/list", Handler: WorkspaceListHandler(ctx)},
		{Method: "POST", Path: "/api/v1/workspace/save", Handler: WorkspaceSaveHandler(ctx)},
		{Method: "POST", Path: "/api/v1/workspace/delete", Handler: WorkspaceDeleteHandler(ctx)},
		// 资产管理
		{Method: "POST", Path: "/api/v1/asset/list", Handler: AssetListHandler(ctx)},
		{Method: "POST", Path: "/api/v1/asset/stat", Handler: AssetStatHandler(ctx)},
		{Method: "POST", Path: "/api/v1/asset/delete", Handler: AssetDeleteHandler(ctx)},
		{Method: "POST", Path: "/api/v1/asset/batchDelete", Handler: AssetBatchDeleteHandler(ctx)},
		{Method: "POST", Path: "/api/v1/asset/history", Handler: AssetHistoryHandler(ctx)},
		// 任务管理
		{Method: "POST", Path: "/api/v1/task/list", Handler: MainTaskListHandler(ctx)},
		{Method: "POST", Path: "/api/v1/task/create", Handler: MainTaskCreateHandler(ctx)},
		{Method: "POST", Path: "/api/v1/task/delete", Handler: MainTaskDeleteHandler(ctx)},
		{Method: "POST", Path: "/api/v1/task/batchDelete", Handler: MainTaskBatchDeleteHandler(ctx)},
		{Method: "POST", Path: "/api/v1/task/retry", Handler: MainTaskRetryHandler(ctx)},
		{Method: "POST", Path: "/api/v1/task/start", Handler: MainTaskStartHandler(ctx)},
		{Method: "POST", Path: "/api/v1/task/pause", Handler: MainTaskPauseHandler(ctx)},
		{Method: "POST", Path: "/api/v1/task/resume", Handler: MainTaskResumeHandler(ctx)},
		{Method: "POST", Path: "/api/v1/task/stop", Handler: MainTaskStopHandler(ctx)},
		{Method: "POST", Path: "/api/v1/task/profile/list", Handler: TaskProfileListHandler(ctx)},
		{Method: "POST", Path: "/api/v1/task/profile/save", Handler: TaskProfileSaveHandler(ctx)},
		{Method: "POST", Path: "/api/v1/task/profile/delete", Handler: TaskProfileDeleteHandler(ctx)},
		// 漏洞管理
		{Method: "POST", Path: "/api/v1/vul/list", Handler: VulListHandler(ctx)},
		{Method: "POST", Path: "/api/v1/vul/delete", Handler: VulDeleteHandler(ctx)},
		{Method: "POST", Path: "/api/v1/vul/batchDelete", Handler: VulBatchDeleteHandler(ctx)},
		// Worker管理
		{Method: "POST", Path: "/api/v1/worker/list", Handler: WorkerListHandler(ctx)},
		// 在线API搜索
		{Method: "POST", Path: "/api/v1/onlineapi/search", Handler: OnlineSearchHandler(ctx)},
		{Method: "POST", Path: "/api/v1/onlineapi/import", Handler: OnlineImportHandler(ctx)},
		{Method: "POST", Path: "/api/v1/onlineapi/config/list", Handler: APIConfigListHandler(ctx)},
		{Method: "POST", Path: "/api/v1/onlineapi/config/save", Handler: APIConfigSaveHandler(ctx)},
		// POC标签映射
		{Method: "POST", Path: "/api/v1/poc/tagmapping/list", Handler: TagMappingListHandler(ctx)},
		{Method: "POST", Path: "/api/v1/poc/tagmapping/save", Handler: TagMappingSaveHandler(ctx)},
		{Method: "POST", Path: "/api/v1/poc/tagmapping/delete", Handler: TagMappingDeleteHandler(ctx)},
		// 自定义POC
		{Method: "POST", Path: "/api/v1/poc/custom/list", Handler: CustomPocListHandler(ctx)},
		{Method: "POST", Path: "/api/v1/poc/custom/save", Handler: CustomPocSaveHandler(ctx)},
		{Method: "POST", Path: "/api/v1/poc/custom/delete", Handler: CustomPocDeleteHandler(ctx)},
		{Method: "POST", Path: "/api/v1/poc/custom/batchImport", Handler: CustomPocBatchImportHandler(ctx)},
		{Method: "POST", Path: "/api/v1/poc/custom/clearAll", Handler: CustomPocClearAllHandler(ctx)},
		// Nuclei默认模板
		{Method: "POST", Path: "/api/v1/poc/nuclei/templates", Handler: NucleiTemplateListHandler(ctx)},
		{Method: "POST", Path: "/api/v1/poc/nuclei/categories", Handler: NucleiTemplateCategoriesHandler(ctx)},
		{Method: "POST", Path: "/api/v1/poc/nuclei/sync", Handler: NucleiTemplateSyncHandler(ctx)},
		{Method: "POST", Path: "/api/v1/poc/nuclei/updateEnabled", Handler: NucleiTemplateUpdateEnabledHandler(ctx)},
		{Method: "POST", Path: "/api/v1/poc/nuclei/detail", Handler: NucleiTemplateDetailHandler(ctx)},
		// 指纹管理
		{Method: "POST", Path: "/api/v1/fingerprint/list", Handler: FingerprintListHandler(ctx)},
		{Method: "POST", Path: "/api/v1/fingerprint/save", Handler: FingerprintSaveHandler(ctx)},
		{Method: "POST", Path: "/api/v1/fingerprint/delete", Handler: FingerprintDeleteHandler(ctx)},
		{Method: "POST", Path: "/api/v1/fingerprint/categories", Handler: FingerprintCategoriesHandler(ctx)},
		{Method: "POST", Path: "/api/v1/fingerprint/sync", Handler: FingerprintSyncHandler(ctx)},
		{Method: "POST", Path: "/api/v1/fingerprint/updateEnabled", Handler: FingerprintUpdateEnabledHandler(ctx)},
		{Method: "POST", Path: "/api/v1/fingerprint/import", Handler: FingerprintImportHandler(ctx)},
		{Method: "POST", Path: "/api/v1/fingerprint/importFromFile", Handler: FingerprintImportFromFileHandler(ctx)},
		{Method: "POST", Path: "/api/v1/fingerprint/clearCustom", Handler: FingerprintClearCustomHandler(ctx)},
		{Method: "POST", Path: "/api/v1/fingerprint/validate", Handler: FingerprintValidateHandler(ctx)},
		{Method: "POST", Path: "/api/v1/fingerprint/batchValidate", Handler: FingerprintBatchValidateHandler(ctx)},
		// POC验证
		{Method: "POST", Path: "/api/v1/poc/custom/validate", Handler: PocValidateHandler(ctx)},
		{Method: "POST", Path: "/api/v1/poc/batchValidate", Handler: PocBatchValidateHandler(ctx)},
		{Method: "POST", Path: "/api/v1/poc/queryResult", Handler: PocValidationResultQueryHandler(ctx)},
	}

	// 为每个路由包装认证中间件
	for i := range authRoutes {
		originalHandler := authRoutes[i].Handler
		authRoutes[i].Handler = func(w http.ResponseWriter, r *http.Request) {
			authMiddleware.Handle(http.HandlerFunc(originalHandler)).ServeHTTP(w, r)
		}
	}

	server.AddRoutes(authRoutes)
}
