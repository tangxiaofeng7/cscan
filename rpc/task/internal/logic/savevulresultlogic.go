package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveVulResultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveVulResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveVulResultLogic {
	return &SaveVulResultLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存漏洞结果
func (l *SaveVulResultLogic) SaveVulResult(in *pb.SaveVulResultReq) (*pb.SaveVulResultResp, error) {
	// todo: add your logic here and delete this line

	return &pb.SaveVulResultResp{}, nil
}
