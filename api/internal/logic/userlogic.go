package logic

import (
	"context"

	"cscan/api/internal/svc"
	"cscan/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

type UserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserListLogic {
	return &UserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserListLogic) UserList(req *types.PageReq) (resp *types.UserListResp, err error) {
	filter := bson.M{}

	total, err := l.svcCtx.UserModel.Count(l.ctx, filter)
	if err != nil {
		return &types.UserListResp{Code: 500, Msg: "查询失败"}, nil
	}

	users, err := l.svcCtx.UserModel.Find(l.ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return &types.UserListResp{Code: 500, Msg: "查询失败"}, nil
	}

	list := make([]types.UserInfo, 0, len(users))
	for _, u := range users {
		list = append(list, types.UserInfo{
			Id:       u.Id.Hex(),
			Username: u.Username,
			Role:     u.Role,
			Status:   u.Status,
		})
	}

	return &types.UserListResp{
		Code:  0,
		Msg:   "success",
		Total: int(total),
		List:  list,
	}, nil
}
