package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTemplatesByTagsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTemplatesByTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTemplatesByTagsLogic {
	return &GetTemplatesByTagsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据标签获取模板
func (l *GetTemplatesByTagsLogic) GetTemplatesByTags(in *pb.GetTemplatesByTagsReq) (*pb.GetTemplatesByTagsResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetTemplatesByTagsResp{}, nil
}
