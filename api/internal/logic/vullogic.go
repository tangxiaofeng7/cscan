package logic

import (
	"context"
	"strconv"

	"cscan/api/internal/svc"
	"cscan/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

type VulListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVulListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VulListLogic {
	return &VulListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VulListLogic) VulList(req *types.VulListReq, workspaceId string) (resp *types.VulListResp, err error) {
	vulModel := l.svcCtx.GetVulModel(workspaceId)

	// 构建查询条件
	filter := bson.M{}
	if req.Authority != "" {
		filter["authority"] = bson.M{"$regex": req.Authority, "$options": "i"}
	}
	if req.Severity != "" {
		filter["severity"] = req.Severity
	}
	if req.Source != "" {
		filter["source"] = req.Source
	}

	// 查询总数
	total, err := vulModel.Count(l.ctx, filter)
	if err != nil {
		return &types.VulListResp{Code: 500, Msg: "查询失败"}, nil
	}

	// 查询列表
	vuls, err := vulModel.Find(l.ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return &types.VulListResp{Code: 500, Msg: "查询失败"}, nil
	}

	// 转换响应
	list := make([]types.Vul, 0, len(vuls))
	for _, v := range vuls {
		list = append(list, types.Vul{
			Id:         v.Id.Hex(),
			Authority:  v.Authority,
			Url:        v.Url,
			PocFile:    v.PocFile,
			Source:     v.Source,
			Severity:   v.Severity,
			Result:     v.Result,
			CreateTime: v.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.VulListResp{
		Code:  0,
		Msg:   "success",
		Total: int(total),
		List:  list,
	}, nil
}


// VulLogic 漏洞管理逻辑
type VulLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVulLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VulLogic {
	return &VulLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VulLogic) VulDelete(req *types.VulDeleteReq, workspaceId string) (resp *types.BaseResp, err error) {
	vulModel := l.svcCtx.GetVulModel(workspaceId)
	if err := vulModel.Delete(l.ctx, req.Id); err != nil {
		return &types.BaseResp{Code: 500, Msg: "删除失败: " + err.Error()}, nil
	}
	return &types.BaseResp{Code: 0, Msg: "删除成功"}, nil
}

func (l *VulLogic) VulBatchDelete(req *types.VulBatchDeleteReq, workspaceId string) (resp *types.BaseResp, err error) {
	vulModel := l.svcCtx.GetVulModel(workspaceId)
	deleted, err := vulModel.BatchDelete(l.ctx, req.Ids)
	if err != nil {
		return &types.BaseResp{Code: 500, Msg: "删除失败: " + err.Error()}, nil
	}
	return &types.BaseResp{Code: 0, Msg: "成功删除 " + strconv.FormatInt(deleted, 10) + " 条记录"}, nil
}
