package user

import (
	"net/http"

	"cscan/api/internal/logic"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/pkg/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// LoginHandler 用户登录
func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// UserListHandler 用户列表
func UserListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PageReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewUserListLogic(r.Context(), svcCtx)
		resp, err := l.UserList(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// UserCreateHandler 创建用户
func UserCreateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserCreateReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewUserCreateLogic(r.Context(), svcCtx)
		resp, err := l.UserCreate(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// UserUpdateHandler 更新用户
func UserUpdateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserUpdateReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewUserUpdateLogic(r.Context(), svcCtx)
		resp, err := l.UserUpdate(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// UserDeleteHandler 删除用户
func UserDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewUserDeleteLogic(r.Context(), svcCtx)
		resp, err := l.UserDelete(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// UserResetPasswordHandler 重置用户密码
func UserResetPasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserResetPasswordReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewUserResetPasswordLogic(r.Context(), svcCtx)
		resp, err := l.UserResetPassword(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// SaveScanConfigHandler 保存用户扫描配置
func SaveScanConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SaveScanConfigReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewScanConfigLogic(r.Context(), svcCtx)
		resp, err := l.SaveScanConfig(r, &req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// GetScanConfigHandler 获取用户扫描配置
func GetScanConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewScanConfigLogic(r.Context(), svcCtx)
		resp, err := l.GetScanConfig(r)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}
