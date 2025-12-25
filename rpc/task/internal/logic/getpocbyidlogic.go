package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPocByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPocByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPocByIdLogic {
	return &GetPocByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据ID获取POC内容
func (l *GetPocByIdLogic) GetPocById(in *pb.GetPocByIdReq) (*pb.GetPocByIdResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetPocByIdResp{}, nil
}
