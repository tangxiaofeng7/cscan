package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RequestResourceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRequestResourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RequestResourceLogic {
	return &RequestResourceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 请求资源文件
func (l *RequestResourceLogic) RequestResource(in *pb.RequestResourceReq) (*pb.RequestResourceResp, error) {
	// todo: add your logic here and delete this line

	return &pb.RequestResourceResp{}, nil
}
