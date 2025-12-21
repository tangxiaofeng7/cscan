package report

import (
	"encoding/json"
	"net/http"
	"net/url"

	"cscan/api/internal/logic"
	"cscan/api/internal/middleware"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"
)

func ReportDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReportDetailReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.ReportDetailResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewReportDetailLogic(r.Context(), svcCtx)
		resp, _ := l.ReportDetail(&req, workspaceId)
		httpResult(w, resp)
	}
}

func ReportExportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReportExportReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, &types.BaseResp{Code: 400, Msg: "参数错误"})
			return
		}

		workspaceId := middleware.GetWorkspaceId(r.Context())
		l := logic.NewReportExportLogic(r.Context(), svcCtx)
		data, filename, err := l.ReportExport(&req, workspaceId)
		if err != nil {
			httpResult(w, &types.BaseResp{Code: 500, Msg: err.Error()})
			return
		}

		// 设置响应头
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(filename))
		w.Write(data)
	}
}

func httpResult(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
