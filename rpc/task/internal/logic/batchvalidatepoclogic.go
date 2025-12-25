package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchValidatePocLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchValidatePocLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchValidatePocLogic {
	return &BatchValidatePocLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量POC验证
func (l *BatchValidatePocLogic) BatchValidatePoc(in *pb.BatchValidatePocReq) (*pb.BatchValidatePocResp, error) {
	// todo: add your logic here and delete this line

	return &pb.BatchValidatePocResp{}, nil
}
