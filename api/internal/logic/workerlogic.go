package logic

import (
	"context"
	"encoding/json"
	"fmt"
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
	IP                 string  `json:"ip"`
	CPULoad            float64 `json:"cpuLoad"`
	MemUsed            float64 `json:"memUsed"`
	TaskStartedNumber  int     `json:"taskStartedNumber"`
	TaskExecutedNumber int     `json:"taskExecutedNumber"`
	UpdateTime         string  `json:"updateTime"`
}

func (l *WorkerListLogic) WorkerList() (resp *types.WorkerListResp, err error) {
	rdb := l.svcCtx.RedisClient

	// 发送查询请求，通知所有Worker立即上报状态
	rdb.Publish(l.ctx, "cscan:worker:query", "refresh")

	// 等待Worker响应（最多等待500毫秒）
	time.Sleep(500 * time.Millisecond)

	// 从Redis获取Worker状态
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
		// 心跳间隔30秒，如果60秒内有更新则认为在线
		workerStatus := "offline"
		if status.UpdateTime != "" {
			loc := time.Local
			updateTime, err := time.ParseInLocation("2006-01-02 15:04:05", status.UpdateTime, loc)
			if err == nil {
				elapsed := time.Since(updateTime)
				l.Logger.Infof("Worker %s: updateTime=%s, elapsed=%v", status.WorkerName, status.UpdateTime, elapsed)
				if elapsed < 60*time.Second {
					workerStatus = "running"
				}
			} else {
				l.Logger.Errorf("Parse time error: %v", err)
			}
		}

		// 计算正在执行的任务数
		runningCount := status.TaskStartedNumber - status.TaskExecutedNumber
		if runningCount < 0 {
			runningCount = 0
		}

		list = append(list, types.Worker{
			Name:         status.WorkerName,
			IP:           status.IP,
			CPULoad:      status.CPULoad,
			MemUsed:      status.MemUsed,
			TaskCount:    status.TaskExecutedNumber,
			RunningCount: runningCount,
			Status:       workerStatus,
			UpdateTime:   status.UpdateTime,
		})
	}

	return &types.WorkerListResp{
		Code: 0,
		Msg:  "success",
		List: list,
	}, nil
}


// WorkerDeleteLogic Worker删除逻辑
type WorkerDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkerDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkerDeleteLogic {
	return &WorkerDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WorkerDeleteLogic) WorkerDelete(req *types.WorkerDeleteReq) (resp *types.WorkerDeleteResp, err error) {
	if req.Name == "" {
		return &types.WorkerDeleteResp{Code: 400, Msg: "Worker名称不能为空"}, nil
	}

	rdb := l.svcCtx.RedisClient

	// 1. 通过Pub/Sub发送停止命令（立即通知在线Worker）
	stopMsg := fmt.Sprintf(`{"action":"stop","workerName":"%s"}`, req.Name)
	rdb.Publish(l.ctx, "cscan:worker:control", stopMsg)
	l.Logger.Infof("[WorkerDelete] Sent stop command to worker: %s", req.Name)

	// 2. 删除Worker状态数据
	workerKey := fmt.Sprintf("worker:%s", req.Name)
	rdb.Del(l.ctx, workerKey)

	// 3. 删除控制信号（避免新启动的同名Worker被误停止）
	ctrlKey := fmt.Sprintf("worker_ctrl:%s", req.Name)
	rdb.Del(l.ctx, ctrlKey)

	l.Logger.Infof("[WorkerDelete] Deleted worker data: %s", req.Name)

	return &types.WorkerDeleteResp{Code: 0, Msg: "Worker已删除，停止信号已发送"}, nil
}

// WorkerRenameLogic Worker重命名逻辑
type WorkerRenameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWorkerRenameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WorkerRenameLogic {
	return &WorkerRenameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WorkerRenameLogic) WorkerRename(req *types.WorkerRenameReq) (resp *types.WorkerRenameResp, err error) {
	if req.OldName == "" || req.NewName == "" {
		return &types.WorkerRenameResp{Code: 400, Msg: "Worker名称不能为空"}, nil
	}

	if req.OldName == req.NewName {
		return &types.WorkerRenameResp{Code: 400, Msg: "新旧名称相同"}, nil
	}

	rdb := l.svcCtx.RedisClient

	// 1. 获取原Worker状态数据
	oldKey := fmt.Sprintf("worker:%s", req.OldName)
	data, err := rdb.Get(l.ctx, oldKey).Result()
	if err != nil {
		return &types.WorkerRenameResp{Code: 404, Msg: "Worker不存在"}, nil
	}

	// 2. 检查新名称是否已存在
	newKey := fmt.Sprintf("worker:%s", req.NewName)
	exists, _ := rdb.Exists(l.ctx, newKey).Result()
	if exists > 0 {
		return &types.WorkerRenameResp{Code: 400, Msg: "新名称已被使用"}, nil
	}

	// 3. 更新状态数据中的workerName
	var status map[string]interface{}
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return &types.WorkerRenameResp{Code: 500, Msg: "数据解析失败"}, nil
	}
	status["workerName"] = req.NewName

	// 4. 保存到新key
	newData, _ := json.Marshal(status)
	rdb.Set(l.ctx, newKey, newData, 10*time.Minute)

	// 5. 删除旧key
	rdb.Del(l.ctx, oldKey)

	// 6. 发送重命名命令给Worker（让Worker更新自己的名称）
	renameMsg := fmt.Sprintf(`{"action":"rename","workerName":"%s","newName":"%s"}`, req.OldName, req.NewName)
	rdb.Publish(l.ctx, "cscan:worker:control", renameMsg)

	l.Logger.Infof("[WorkerRename] Renamed worker from %s to %s", req.OldName, req.NewName)

	return &types.WorkerRenameResp{Code: 0, Msg: "重命名成功"}, nil
}
