package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWorkerConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetWorkerConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWorkerConfigLogic {
	return &GetWorkerConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取Worker配置
func (l *GetWorkerConfigLogic) GetWorkerConfig(in *pb.GetWorkerConfigReq) (*pb.GetWorkerConfigResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetWorkerConfigResp{}, nil
}
