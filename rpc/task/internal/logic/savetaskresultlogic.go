package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveTaskResultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveTaskResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveTaskResultLogic {
	return &SaveTaskResultLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存任务结果
func (l *SaveTaskResultLogic) SaveTaskResult(in *pb.SaveTaskResultReq) (*pb.SaveTaskResultResp, error) {
	// todo: add your logic here and delete this line

	return &pb.SaveTaskResultResp{}, nil
}
