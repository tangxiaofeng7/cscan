package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHttpServiceMappingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHttpServiceMappingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHttpServiceMappingsLogic {
	return &GetHttpServiceMappingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取HTTP服务映射
func (l *GetHttpServiceMappingsLogic) GetHttpServiceMappings(in *pb.GetHttpServiceMappingsReq) (*pb.GetHttpServiceMappingsResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetHttpServiceMappingsResp{}, nil
}
