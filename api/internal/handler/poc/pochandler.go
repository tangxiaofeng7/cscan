package poc

import (
	"net/http"

	"cscan/api/internal/logic"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/pkg/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// TagMappingListHandler POC标签映射列表
func TagMappingListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewTagMappingListLogic(r.Context(), svcCtx)
		resp, err := l.TagMappingList()
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// TagMappingSaveHandler 保存POC标签映射
func TagMappingSaveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TagMappingSaveReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewTagMappingSaveLogic(r.Context(), svcCtx)
		resp, err := l.TagMappingSave(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// TagMappingDeleteHandler 删除POC标签映射
func TagMappingDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TagMappingDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewTagMappingDeleteLogic(r.Context(), svcCtx)
		resp, err := l.TagMappingDelete(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// CustomPocListHandler 自定义POC列表
func CustomPocListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CustomPocListReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewCustomPocListLogic(r.Context(), svcCtx)
		resp, err := l.CustomPocList(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// CustomPocSaveHandler 保存自定义POC
func CustomPocSaveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CustomPocSaveReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewCustomPocSaveLogic(r.Context(), svcCtx)
		resp, err := l.CustomPocSave(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// CustomPocDeleteHandler 删除自定义POC
func CustomPocDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CustomPocDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewCustomPocDeleteLogic(r.Context(), svcCtx)
		resp, err := l.CustomPocDelete(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// CustomPocBatchImportHandler 批量导入自定义POC
func CustomPocBatchImportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CustomPocBatchImportReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewCustomPocBatchImportLogic(r.Context(), svcCtx)
		resp, err := l.CustomPocBatchImport(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// CustomPocClearAllHandler 清空所有自定义POC
func CustomPocClearAllHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewCustomPocClearAllLogic(r.Context(), svcCtx)
		resp, err := l.CustomPocClearAll()
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// NucleiTemplateListHandler Nuclei模板列表
func NucleiTemplateListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.NucleiTemplateListReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewNucleiTemplateListLogic(r.Context(), svcCtx)
		resp, err := l.NucleiTemplateList(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// NucleiTemplateCategoriesHandler Nuclei模板分类
func NucleiTemplateCategoriesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewNucleiTemplateCategoriesLogic(r.Context(), svcCtx)
		resp, err := l.NucleiTemplateCategories()
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// NucleiTemplateSyncHandler 同步Nuclei模板
func NucleiTemplateSyncHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Force bool `json:"force"`
		}
		httpx.Parse(r, &req)

		if req.Force {
			svcCtx.NucleiTemplateModel.DeleteAll(r.Context())
		}

		go svcCtx.SyncNucleiTemplates()
		httpx.OkJson(w, &types.BaseResp{Code: 0, Msg: "模板同步已开始，请稍后刷新查看"})
	}
}

// NucleiTemplateUpdateEnabledHandler 更新Nuclei模板启用状态
func NucleiTemplateUpdateEnabledHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.NucleiTemplateUpdateEnabledReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewNucleiTemplateUpdateEnabledLogic(r.Context(), svcCtx)
		resp, err := l.UpdateEnabled(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// NucleiTemplateDetailHandler Nuclei模板详情
func NucleiTemplateDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.NucleiTemplateDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewNucleiTemplateDetailLogic(r.Context(), svcCtx)
		resp, err := l.GetDetail(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// PocValidateHandler POC验证
func PocValidateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PocValidateReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewPocValidateLogic(r.Context(), svcCtx)
		resp, err := l.PocValidate(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// PocBatchValidateHandler 批量POC验证
func PocBatchValidateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PocBatchValidateReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewPocBatchValidateLogic(r.Context(), svcCtx)
		resp, err := l.PocBatchValidate(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// PocValidationResultQueryHandler 查询POC验证结果
func PocValidationResultQueryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PocValidationResultQueryReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewPocValidationResultQueryLogic(r.Context(), svcCtx)
		resp, err := l.PocValidationResultQuery(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}
