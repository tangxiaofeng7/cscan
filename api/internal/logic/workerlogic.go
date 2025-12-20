package logic

import (
	"context"
	"encoding/json"
	"time"

	"cscan/api/internal/svc"
	"cscan/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WorkerListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkerListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkerListLogic {
	return &WorkerListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type WorkerStatus struct {
	WorkerName         string  `json:"workerName"`
	CPULoad            float64 `json:"cpuLoad"`
	MemUsed            float64 `json:"memUsed"`
	TaskStartedNumber  int     `json:"taskStartedNumber"`
	TaskExecutedNumber int     `json:"taskExecutedNumber"`
	UpdateTime         string  `json:"updateTime"`
}

func (l *WorkerListLogic) WorkerList() (resp *types.WorkerListResp, err error) {
	// 从Redis获取Worker状态
	rdb := l.svcCtx.RedisClient
	keys, err := rdb.Keys(l.ctx, "worker:*").Result()
	if err != nil {
		return &types.WorkerListResp{Code: 500, Msg: "查询失败"}, nil
	}

	list := make([]types.Worker, 0, len(keys))
	for _, key := range keys {
		data, err := rdb.Get(l.ctx, key).Result()
		if err != nil {
			continue
		}

		var status WorkerStatus
		if err := json.Unmarshal([]byte(data), &status); err != nil {
			continue
		}

		// 根据最后更新时间判断在线状态
		// Worker心跳间隔是30秒，如果超过45秒没有心跳，认为离线
		workerStatus := "offline"
		if status.UpdateTime != "" {
			// 使用本地时区解析时间
			loc := time.Local
			updateTime, err := time.ParseInLocation("2006-01-02 15:04:05", status.UpdateTime, loc)
			if err == nil {
				elapsed := time.Since(updateTime)
				l.Logger.Infof("Worker %s: updateTime=%s, elapsed=%v", status.WorkerName, status.UpdateTime, elapsed)
				if elapsed < 45*time.Second {
					workerStatus = "running"
				}
			} else {
				l.Logger.Errorf("Parse time error: %v", err)
			}
		}

		list = append(list, types.Worker{
			Name:       status.WorkerName,
			CPULoad:    status.CPULoad,
			MemUsed:    status.MemUsed,
			TaskCount:  status.TaskStartedNumber,
			Status:     workerStatus,
			UpdateTime: status.UpdateTime,
		})
	}

	return &types.WorkerListResp{
		Code: 0,
		Msg:  "success",
		List: list,
	}, nil
}
