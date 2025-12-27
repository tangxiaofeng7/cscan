package logic

import (
	"context"
	"net/http"
	"time"

	"cscan/api/internal/middleware"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/model"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

type UserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserListLogic {
	return &UserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserListLogic) UserList(req *types.PageReq) (resp *types.UserListResp, err error) {
	filter := bson.M{}

	total, err := l.svcCtx.UserModel.Count(l.ctx, filter)
	if err != nil {
		return &types.UserListResp{Code: 500, Msg: "查询失败"}, nil
	}

	users, err := l.svcCtx.UserModel.Find(l.ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return &types.UserListResp{Code: 500, Msg: "查询失败"}, nil
	}

	list := make([]types.UserInfo, 0, len(users))
	for _, u := range users {
		list = append(list, types.UserInfo{
			Id:       u.Id.Hex(),
			Username: u.Username,
			Status:   u.Status,
		})
	}

	return &types.UserListResp{
		Code:  0,
		Msg:   "success",
		Total: int(total),
		List:  list,
	}, nil
}

// UserCreateLogic 创建用户逻辑
type UserCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserCreateLogic {
	return &UserCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserCreateLogic) UserCreate(req *types.UserCreateReq) (resp *types.BaseResp, err error) {
	// 检查用户名是否已存在
	exists, err := l.svcCtx.UserModel.FindByUsername(l.ctx, req.Username)
	if err != nil {
		logx.Errorf("查询用户失败: %v", err)
		return &types.BaseResp{Code: 500, Msg: "系统错误"}, nil
	}
	if exists != nil {
		return &types.BaseResp{Code: 400, Msg: "用户名已存在"}, nil
	}

	// 创建用户
	user := &model.User{
		Username: req.Username,
		Password: req.Password, // 在model层会自动MD5加密
		Status:   req.Status,
	}

	err = l.svcCtx.UserModel.Insert(l.ctx, user)
	if err != nil {
		logx.Errorf("创建用户失败: %v", err)
		return &types.BaseResp{Code: 500, Msg: "创建用户失败"}, nil
	}

	return &types.BaseResp{Code: 0, Msg: "创建成功"}, nil
}

// UserUpdateLogic 更新用户逻辑
type UserUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserUpdateLogic {
	return &UserUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserUpdateLogic) UserUpdate(req *types.UserUpdateReq) (resp *types.BaseResp, err error) {
	// 检查用户是否存在
	user, err := l.svcCtx.UserModel.FindById(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("查询用户失败: %v", err)
		return &types.BaseResp{Code: 500, Msg: "系统错误"}, nil
	}
	if user == nil {
		return &types.BaseResp{Code: 404, Msg: "用户不存在"}, nil
	}

	// 如果修改用户名，检查是否重复
	if req.Username != user.Username {
		exists, err := l.svcCtx.UserModel.FindByUsername(l.ctx, req.Username)
		if err != nil {
			logx.Errorf("查询用户失败: %v", err)
			return &types.BaseResp{Code: 500, Msg: "系统错误"}, nil
		}
		if exists != nil {
			return &types.BaseResp{Code: 400, Msg: "用户名已存在"}, nil
		}
	}

	// 更新用户信息
	updateData := bson.M{
		"username": req.Username,
		"status":   req.Status,
		"update_time": time.Now(),
	}

	err = l.svcCtx.UserModel.UpdateById(l.ctx, req.Id, updateData)
	if err != nil {
		logx.Errorf("更新用户失败: %v", err)
		return &types.BaseResp{Code: 500, Msg: "更新用户失败"}, nil
	}

	return &types.BaseResp{Code: 0, Msg: "更新成功"}, nil
}

// UserDeleteLogic 删除用户逻辑
type UserDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserDeleteLogic {
	return &UserDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserDeleteLogic) UserDelete(req *types.UserDeleteReq) (resp *types.BaseResp, err error) {
	// 检查用户是否存在
	user, err := l.svcCtx.UserModel.FindById(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("查询用户失败: %v", err)
		return &types.BaseResp{Code: 500, Msg: "系统错误"}, nil
	}
	if user == nil {
		return &types.BaseResp{Code: 404, Msg: "用户不存在"}, nil
	}

	// 禁止删除 admin 账号
	if user.Username == "admin" {
		return &types.BaseResp{Code: 400, Msg: "admin 账号不允许删除"}, nil
	}

	// 删除用户
	err = l.svcCtx.UserModel.DeleteById(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("删除用户失败: %v", err)
		return &types.BaseResp{Code: 500, Msg: "删除用户失败"}, nil
	}

	return &types.BaseResp{Code: 0, Msg: "删除成功"}, nil
}

// UserResetPasswordLogic 重置密码逻辑
type UserResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserResetPasswordLogic {
	return &UserResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserResetPasswordLogic) UserResetPassword(req *types.UserResetPasswordReq) (resp *types.BaseResp, err error) {
	// 检查用户是否存在
	user, err := l.svcCtx.UserModel.FindById(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("查询用户失败: %v", err)
		return &types.BaseResp{Code: 500, Msg: "系统错误"}, nil
	}
	if user == nil {
		return &types.BaseResp{Code: 404, Msg: "用户不存在"}, nil
	}

	// 重置密码
	err = l.svcCtx.UserModel.UpdatePassword(l.ctx, req.Id, req.NewPassword)
	if err != nil {
		logx.Errorf("重置密码失败: %v", err)
		return &types.BaseResp{Code: 500, Msg: "重置密码失败"}, nil
	}

	return &types.BaseResp{Code: 0, Msg: "密码重置成功"}, nil
}


// ScanConfigLogic 扫描配置逻辑
type ScanConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewScanConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScanConfigLogic {
	return &ScanConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ScanConfigLogic) SaveScanConfig(r *http.Request, req *types.SaveScanConfigReq) (resp *types.BaseResp, err error) {
	// 从请求上下文获取用户ID
	userId := middleware.GetUserId(r.Context())
	if userId == "" {
		return &types.BaseResp{Code: 401, Msg: "未登录"}, nil
	}

	err = l.svcCtx.UserModel.UpdateScanConfig(l.ctx, userId, req.Config)
	if err != nil {
		logx.Errorf("保存扫描配置失败: %v", err)
		return &types.BaseResp{Code: 500, Msg: "保存失败"}, nil
	}

	return &types.BaseResp{Code: 0, Msg: "保存成功"}, nil
}

func (l *ScanConfigLogic) GetScanConfig(r *http.Request) (resp *types.GetScanConfigResp, err error) {
	// 从请求上下文获取用户ID
	userId := middleware.GetUserId(r.Context())
	if userId == "" {
		return &types.GetScanConfigResp{Code: 401, Msg: "未登录"}, nil
	}

	config, err := l.svcCtx.UserModel.GetScanConfig(l.ctx, userId)
	if err != nil {
		logx.Errorf("获取扫描配置失败: %v", err)
		return &types.GetScanConfigResp{Code: 500, Msg: "获取失败"}, nil
	}

	return &types.GetScanConfigResp{Code: 0, Msg: "success", Config: config}, nil
}
