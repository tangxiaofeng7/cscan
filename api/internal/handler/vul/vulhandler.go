package vul

import (
	"net/http"

	"cscan/api/internal/logic"
	"cscan/api/internal/middleware"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/pkg/response"
	"cscan/pkg/xerr"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// VulListHandler 漏洞列表
func VulListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VulListReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewVulListLogic(r.Context(), svcCtx)
		resp, err := l.VulList(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// VulDetailHandler 漏洞详情 
func VulDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VulDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}
		if req.Id == "" {
			response.Error(w, xerr.NewParamError("漏洞ID不能为空"))
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewVulDetailLogic(r.Context(), svcCtx)
		resp, err := l.VulDetail(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// VulDeleteHandler 删除漏洞
func VulDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VulDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}
		if req.Id == "" {
			response.Error(w, xerr.NewParamError("ID不能为空"))
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewVulLogic(r.Context(), svcCtx)
		resp, err := l.VulDelete(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// VulBatchDeleteHandler 批量删除漏洞
func VulBatchDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VulBatchDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}
		if len(req.Ids) == 0 {
			response.Error(w, xerr.NewParamError("请选择要删除的漏洞"))
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewVulLogic(r.Context(), svcCtx)
		resp, err := l.VulBatchDelete(&req, workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// VulStatHandler 漏洞统计
func VulStatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewVulStatLogic(r.Context(), svcCtx)
		resp, err := l.VulStat(workspaceId)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}
