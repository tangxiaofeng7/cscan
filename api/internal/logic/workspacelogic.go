package logic

import (
	"context"

	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/model"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

type WorkspaceListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkspaceListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkspaceListLogic {
	return &WorkspaceListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WorkspaceListLogic) WorkspaceList(req *types.PageReq) (resp *types.WorkspaceListResp, err error) {
	filter := bson.M{}

	total, err := l.svcCtx.WorkspaceModel.Count(l.ctx, filter)
	if err != nil {
		return &types.WorkspaceListResp{Code: 500, Msg: "查询失败"}, nil
	}

	workspaces, err := l.svcCtx.WorkspaceModel.Find(l.ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return &types.WorkspaceListResp{Code: 500, Msg: "查询失败"}, nil
	}

	list := make([]types.Workspace, 0, len(workspaces))
	for _, w := range workspaces {
		list = append(list, types.Workspace{
			Id:          w.Id.Hex(),
			Name:        w.Name,
			Description: w.Description,
			Status:      w.Status,
			CreateTime:  w.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.WorkspaceListResp{
		Code:  0,
		Msg:   "success",
		Total: int(total),
		List:  list,
	}, nil
}

type WorkspaceSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkspaceSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkspaceSaveLogic {
	return &WorkspaceSaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WorkspaceSaveLogic) WorkspaceSave(req *types.WorkspaceSaveReq) (resp *types.BaseResp, err error) {
	if req.Id != "" {
		// 更新
		err = l.svcCtx.WorkspaceModel.Update(l.ctx, req.Id, bson.M{
			"name":        req.Name,
			"description": req.Description,
		})
		if err != nil {
			return &types.BaseResp{Code: 500, Msg: "更新失败"}, nil
		}
		return &types.BaseResp{Code: 0, Msg: "更新成功"}, nil
	}

	// 新增
	workspace := &model.Workspace{
		Name:        req.Name,
		Description: req.Description,
	}
	if err = l.svcCtx.WorkspaceModel.Insert(l.ctx, workspace); err != nil {
		return &types.BaseResp{Code: 500, Msg: "创建失败"}, nil
	}

	return &types.BaseResp{Code: 0, Msg: "创建成功"}, nil
}

type WorkspaceDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkspaceDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkspaceDeleteLogic {
	return &WorkspaceDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WorkspaceDeleteLogic) WorkspaceDelete(req *types.WorkspaceSaveReq) (resp *types.BaseResp, err error) {
	if req.Id == "" {
		return &types.BaseResp{Code: 400, Msg: "ID不能为空"}, nil
	}

	if err = l.svcCtx.WorkspaceModel.Delete(l.ctx, req.Id); err != nil {
		return &types.BaseResp{Code: 500, Msg: "删除失败"}, nil
	}

	return &types.BaseResp{Code: 0, Msg: "删除成功"}, nil
}
