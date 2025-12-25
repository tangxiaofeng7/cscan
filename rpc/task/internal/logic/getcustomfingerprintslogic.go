package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCustomFingerprintsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCustomFingerprintsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCustomFingerprintsLogic {
	return &GetCustomFingerprintsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取自定义指纹
func (l *GetCustomFingerprintsLogic) GetCustomFingerprints(in *pb.GetCustomFingerprintsReq) (*pb.GetCustomFingerprintsResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetCustomFingerprintsResp{}, nil
}
