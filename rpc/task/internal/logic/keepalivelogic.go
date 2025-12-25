package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type KeepAliveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewKeepAliveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KeepAliveLogic {
	return &KeepAliveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Worker心跳
func (l *KeepAliveLogic) KeepAlive(in *pb.KeepAliveReq) (*pb.KeepAliveResp, error) {
	// todo: add your logic here and delete this line

	return &pb.KeepAliveResp{}, nil
}
