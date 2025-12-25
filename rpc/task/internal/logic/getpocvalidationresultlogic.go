package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPocValidationResultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPocValidationResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPocValidationResultLogic {
	return &GetPocValidationResultLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询POC验证结果
func (l *GetPocValidationResultLogic) GetPocValidationResult(in *pb.GetPocValidationResultReq) (*pb.GetPocValidationResultResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetPocValidationResultResp{}, nil
}
