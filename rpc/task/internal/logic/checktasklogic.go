package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckTaskLogic {
	return &CheckTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 检查任务状态
func (l *CheckTaskLogic) CheckTask(in *pb.CheckTaskReq) (*pb.CheckTaskResp, error) {
	// todo: add your logic here and delete this line

	return &pb.CheckTaskResp{}, nil
}
