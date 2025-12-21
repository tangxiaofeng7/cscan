package fingerprint

import (
	"net/http"

	"cscan/api/internal/logic"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/pkg/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// FingerprintListHandler 指纹列表
func FingerprintListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintListReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewFingerprintListLogic(r.Context(), svcCtx)
		resp, err := l.FingerprintList(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// FingerprintSaveHandler 保存指纹
func FingerprintSaveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintSaveReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewFingerprintSaveLogic(r.Context(), svcCtx)
		resp, err := l.FingerprintSave(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// FingerprintDeleteHandler 删除指纹
func FingerprintDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewFingerprintDeleteLogic(r.Context(), svcCtx)
		resp, err := l.FingerprintDelete(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// FingerprintCategoriesHandler 指纹分类
func FingerprintCategoriesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewFingerprintCategoriesLogic(r.Context(), svcCtx)
		resp, err := l.FingerprintCategories()
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// FingerprintSyncHandler 同步指纹
func FingerprintSyncHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintSyncReq
		httpx.Parse(r, &req)

		if req.Force {
			svcCtx.FingerprintModel.DeleteBuiltin(r.Context())
		}

		go svcCtx.SyncWappalyzerFingerprints()
		httpx.OkJson(w, &types.BaseResp{Code: 0, Msg: "指纹同步已开始，请稍后刷新查看"})
	}
}

// FingerprintUpdateEnabledHandler 更新指纹启用状态
func FingerprintUpdateEnabledHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Id      string `json:"id"`
			Enabled bool   `json:"enabled"`
		}
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewFingerprintUpdateEnabledLogic(r.Context(), svcCtx)
		resp, err := l.UpdateEnabled(req.Id, req.Enabled)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// FingerprintImportHandler 导入指纹
func FingerprintImportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintImportReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewFingerprintImportLogic(r.Context(), svcCtx)
		resp, err := l.FingerprintImport(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// FingerprintImportFromFileHandler 从文件导入指纹
func FingerprintImportFromFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintImportFromFileReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewFingerprintImportFromFileLogic(r.Context(), svcCtx)
		resp, err := l.FingerprintImportFromFile(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// FingerprintClearCustomHandler 清空自定义指纹
func FingerprintClearCustomHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintClearCustomReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewFingerprintClearCustomLogic(r.Context(), svcCtx)
		resp, err := l.FingerprintClearCustom(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// FingerprintValidateHandler 验证指纹
func FingerprintValidateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintValidateReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewFingerprintValidateLogic(r.Context(), svcCtx)
		resp, err := l.FingerprintValidate(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// FingerprintBatchValidateHandler 批量验证指纹
func FingerprintBatchValidateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FingerprintBatchValidateReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewFingerprintBatchValidateLogic(r.Context(), svcCtx)
		resp, err := l.FingerprintBatchValidate(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// HttpServiceMappingListHandler HTTP服务映射列表
func HttpServiceMappingListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HttpServiceMappingListReq
		httpx.Parse(r, &req)

		l := logic.NewHttpServiceMappingListLogic(r.Context(), svcCtx)
		resp, err := l.HttpServiceMappingList(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// HttpServiceMappingSaveHandler 保存HTTP服务映射
func HttpServiceMappingSaveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HttpServiceMappingSaveReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewHttpServiceMappingSaveLogic(r.Context(), svcCtx)
		resp, err := l.HttpServiceMappingSave(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// HttpServiceMappingDeleteHandler 删除HTTP服务映射
func HttpServiceMappingDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HttpServiceMappingDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewHttpServiceMappingDeleteLogic(r.Context(), svcCtx)
		resp, err := l.HttpServiceMappingDelete(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}
