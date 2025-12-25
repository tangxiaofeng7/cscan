package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTemplatesByIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTemplatesByIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTemplatesByIdsLogic {
	return &GetTemplatesByIdsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据ID列表批量获取模板内容
func (l *GetTemplatesByIdsLogic) GetTemplatesByIds(in *pb.GetTemplatesByIdsReq) (*pb.GetTemplatesByIdsResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetTemplatesByIdsResp{}, nil
}
