package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateFingerprintLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateFingerprintLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateFingerprintLogic {
	return &ValidateFingerprintLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 验证指纹匹配
func (l *ValidateFingerprintLogic) ValidateFingerprint(in *pb.ValidateFingerprintReq) (*pb.ValidateFingerprintResp, error) {
	// todo: add your logic here and delete this line

	return &pb.ValidateFingerprintResp{}, nil
}
