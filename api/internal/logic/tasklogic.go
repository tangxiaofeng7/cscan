package logic

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"cscan/api/internal/logic/common"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/model"
	"cscan/scheduler"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

type MainTaskListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMainTaskListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MainTaskListLogic {
	return &MainTaskListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MainTaskListLogic) MainTaskList(req *types.MainTaskListReq, workspaceId string) (resp *types.MainTaskListResp, err error) {
	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)

	// 构建查询条件
	filter := bson.M{}
	if req.Name != "" {
		filter["name"] = bson.M{"$regex": req.Name, "$options": "i"}
	}
	if req.Status != "" {
		filter["status"] = req.Status
	}

	// 查询总数
	total, err := taskModel.Count(l.ctx, filter)
	if err != nil {
		return &types.MainTaskListResp{Code: 500, Msg: "查询失败"}, nil
	}

	// 查询列表
	tasks, err := taskModel.Find(l.ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return &types.MainTaskListResp{Code: 500, Msg: "查询失败"}, nil
	}

	// 转换响应
	list := make([]types.MainTask, 0, len(tasks))
	for _, t := range tasks {
		list = append(list, types.MainTask{
			Id:          t.Id.Hex(),
			TaskId:      t.TaskId, // UUID，用于日志查询
			Name:        t.Name,
			Target:      t.Target,
			ProfileId:   t.ProfileId,
			ProfileName: t.ProfileName,
			Status:      t.Status,
			Progress:    t.Progress,
			Result:      t.Result,
			IsCron:      t.IsCron,
			CronRule:    t.CronRule,
			CreateTime:  t.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.MainTaskListResp{
		Code:  0,
		Msg:   "success",
		Total: int(total),
		List:  list,
	}, nil
}

type MainTaskCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMainTaskCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MainTaskCreateLogic {
	return &MainTaskCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MainTaskCreateLogic) MainTaskCreate(req *types.MainTaskCreateReq, workspaceId string) (resp *types.BaseResp, err error) {
	l.Logger.Infof("MainTaskCreate: name=%s, profileId=%s, workspaceId=%s", req.Name, req.ProfileId, workspaceId)

	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)

	// 获取任务配置
	profile, err := l.svcCtx.ProfileModel.FindById(l.ctx, req.ProfileId)
	if err != nil {
		l.Logger.Errorf("MainTaskCreate: profile not found, profileId=%s, error=%v", req.ProfileId, err)
		return &types.BaseResp{Code: 400, Msg: "任务配置不存在"}, nil
	}
	l.Logger.Infof("MainTaskCreate: profile found, name=%s", profile.Name)

	// 构建任务配置，包含目标信息
	taskConfig := map[string]interface{}{
		"target": req.Target,
	}
	// 合并 profile 配置
	if profile.Config != "" {
		var profileConfig map[string]interface{}
		if err := json.Unmarshal([]byte(profile.Config), &profileConfig); err == nil {
			for k, v := range profileConfig {
				taskConfig[k] = v
			}
		}
	}

	// 注入自定义POC和标签映射
	taskConfig = common.InjectPocConfig(l.ctx, l.svcCtx, taskConfig, l.Logger)
	configBytes, _ := json.Marshal(taskConfig)

	// 创建主任务（状态为CREATED，不立即执行）
	taskId := uuid.New().String()
	task := &model.MainTask{
		TaskId:      taskId,
		Name:        req.Name,
		Target:      req.Target,
		ProfileId:   req.ProfileId,
		ProfileName: profile.Name,
		IsCron:      req.IsCron,
		CronRule:    req.CronRule,
		Config:      string(configBytes), // 保存配置用于后续启动
	}

	if err := taskModel.Insert(l.ctx, task); err != nil {
		l.Logger.Errorf("MainTaskCreate: insert failed, taskId=%s, error=%v", taskId, err)
		return &types.BaseResp{Code: 500, Msg: "创建任务失败: " + err.Error()}, nil
	}

	l.Logger.Infof("Task created (not started): taskId=%s, workspaceId=%s", taskId, workspaceId)

	return &types.BaseResp{Code: 0, Msg: "任务创建成功，请手动启动"}, nil
}

type TaskProfileListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTaskProfileListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaskProfileListLogic {
	return &TaskProfileListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TaskProfileListLogic) TaskProfileList() (resp *types.TaskProfileListResp, err error) {
	profiles, err := l.svcCtx.ProfileModel.FindAll(l.ctx)
	if err != nil {
		return &types.TaskProfileListResp{Code: 500, Msg: "查询失败"}, nil
	}

	list := make([]types.TaskProfile, 0, len(profiles))
	for _, p := range profiles {
		list = append(list, types.TaskProfile{
			Id:          p.Id.Hex(),
			Name:        p.Name,
			Description: p.Description,
			Config:      p.Config,
		})
	}

	return &types.TaskProfileListResp{
		Code: 0,
		Msg:  "success",
		List: list,
	}, nil
}

// TaskProfileSaveLogic
type TaskProfileSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTaskProfileSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaskProfileSaveLogic {
	return &TaskProfileSaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TaskProfileSaveLogic) TaskProfileSave(req *types.TaskProfileSaveReq) (resp *types.BaseResp, err error) {
	profile := &model.TaskProfile{
		Name:        req.Name,
		Description: req.Description,
		Config:      req.Config,
	}

	if req.Id != "" {
		// 更新
		err = l.svcCtx.ProfileModel.Update(l.ctx, req.Id, profile)
		if err != nil {
			return &types.BaseResp{Code: 500, Msg: "更新失败"}, nil
		}
	} else {
		// 新增
		err = l.svcCtx.ProfileModel.Insert(l.ctx, profile)
		if err != nil {
			return &types.BaseResp{Code: 500, Msg: "创建失败"}, nil
		}
	}

	return &types.BaseResp{Code: 0, Msg: "保存成功"}, nil
}

// TaskProfileDeleteLogic
type TaskProfileDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTaskProfileDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaskProfileDeleteLogic {
	return &TaskProfileDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TaskProfileDeleteLogic) TaskProfileDelete(req *types.TaskProfileDeleteReq) (resp *types.BaseResp, err error) {
	err = l.svcCtx.ProfileModel.Delete(l.ctx, req.Id)
	if err != nil {
		return &types.BaseResp{Code: 500, Msg: "删除失败"}, nil
	}
	return &types.BaseResp{Code: 0, Msg: "删除成功"}, nil
}

// MainTaskDeleteLogic 单个删除
type MainTaskDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMainTaskDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MainTaskDeleteLogic {
	return &MainTaskDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MainTaskDeleteLogic) MainTaskDelete(req *types.MainTaskDeleteReq, workspaceId string) (resp *types.BaseResp, err error) {
	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)
	err = taskModel.Delete(l.ctx, req.Id)
	if err != nil {
		return &types.BaseResp{Code: 500, Msg: "删除失败"}, nil
	}
	return &types.BaseResp{Code: 0, Msg: "删除成功"}, nil
}

// MainTaskBatchDeleteLogic 批量删除
type MainTaskBatchDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMainTaskBatchDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MainTaskBatchDeleteLogic {
	return &MainTaskBatchDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MainTaskBatchDeleteLogic) MainTaskBatchDelete(req *types.MainTaskBatchDeleteReq, workspaceId string) (resp *types.BaseResp, err error) {
	if len(req.Ids) == 0 {
		return &types.BaseResp{Code: 400, Msg: "请选择要删除的任务"}, nil
	}

	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)
	deleted, err := taskModel.BatchDelete(l.ctx, req.Ids)
	if err != nil {
		return &types.BaseResp{Code: 500, Msg: "删除失败"}, nil
	}
	return &types.BaseResp{Code: 0, Msg: "成功删除 " + strconv.FormatInt(deleted, 10) + " 条任务"}, nil
}


// MainTaskRetryLogic 重新执行任务
type MainTaskRetryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMainTaskRetryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MainTaskRetryLogic {
	return &MainTaskRetryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MainTaskRetryLogic) MainTaskRetry(req *types.MainTaskRetryReq, workspaceId string) (resp *types.BaseResp, err error) {
	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)

	// 获取原任务信息
	task, err := taskModel.FindById(l.ctx, req.Id)
	if err != nil {
		return &types.BaseResp{Code: 400, Msg: "任务不存在"}, nil
	}

	// 获取任务配置
	profile, err := l.svcCtx.ProfileModel.FindById(l.ctx, task.ProfileId)
	if err != nil {
		return &types.BaseResp{Code: 400, Msg: "任务配置不存在"}, nil
	}

	// 生成新的任务ID
	taskId := uuid.New().String()

	// 重置任务状态
	update := bson.M{
		"task_id":  taskId,
		"status":   "PENDING",
		"progress": 0,
		"result":   "",
	}
	if err := taskModel.Update(l.ctx, req.Id, update); err != nil {
		return &types.BaseResp{Code: 500, Msg: "更新任务状态失败"}, nil
	}

	// 构建任务配置
	taskConfig := map[string]interface{}{
		"target": task.Target,
	}
	if profile.Config != "" {
		var profileConfig map[string]interface{}
		if err := json.Unmarshal([]byte(profile.Config), &profileConfig); err == nil {
			for k, v := range profileConfig {
				taskConfig[k] = v
			}
		}
	}

	// 注入自定义POC和标签映射
	taskConfig = common.InjectPocConfig(l.ctx, l.svcCtx, taskConfig, l.Logger)
	configBytes, _ := json.Marshal(taskConfig)

	// 发送任务到消息队列
	schedTask := &scheduler.TaskInfo{
		TaskId:      taskId,
		MainTaskId:  task.Id.Hex(),
		WorkspaceId: workspaceId,
		TaskName:    task.Name,
		Config:      string(configBytes),
		Priority:    1,
	}
	l.Logger.Infof("Retrying task: taskId=%s, workspaceId=%s", taskId, workspaceId)
	if err := l.svcCtx.Scheduler.PushTask(l.ctx, schedTask); err != nil {
		l.Logger.Errorf("push task to queue failed: %v", err)
		return &types.BaseResp{Code: 500, Msg: "任务入队失败"}, nil
	}

	// 保存任务信息到 Redis
	taskInfoKey := "cscan:task:info:" + taskId
	taskInfoData, _ := json.Marshal(map[string]string{
		"workspaceId": workspaceId,
		"mainTaskId":  task.Id.Hex(),
	})
	l.svcCtx.RedisClient.Set(l.ctx, taskInfoKey, taskInfoData, 24*time.Hour)

	return &types.BaseResp{Code: 0, Msg: "任务已重新执行"}, nil
}


// MainTaskStartLogic 启动任务
type MainTaskStartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMainTaskStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MainTaskStartLogic {
	return &MainTaskStartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MainTaskStartLogic) MainTaskStart(req *types.MainTaskControlReq, workspaceId string) (resp *types.BaseResp, err error) {
	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)

	// 获取任务
	task, err := taskModel.FindById(l.ctx, req.Id)
	if err != nil {
		return &types.BaseResp{Code: 400, Msg: "任务不存在"}, nil
	}

	// 检查状态：只有CREATED状态可以启动
	if task.Status != model.TaskStatusCreated {
		return &types.BaseResp{Code: 400, Msg: "只有待启动状态的任务可以启动"}, nil
	}

	// 更新状态为PENDING
	update := bson.M{"status": model.TaskStatusPending}
	if err := taskModel.Update(l.ctx, req.Id, update); err != nil {
		return &types.BaseResp{Code: 500, Msg: "更新任务状态失败"}, nil
	}

	// 发送任务到消息队列
	schedTask := &scheduler.TaskInfo{
		TaskId:      task.TaskId,
		MainTaskId:  task.Id.Hex(),
		WorkspaceId: workspaceId,
		TaskName:    task.Name,
		Config:      task.Config,
		Priority:    1,
	}
	l.Logger.Infof("Starting task: taskId=%s, workspaceId=%s", task.TaskId, workspaceId)
	if err := l.svcCtx.Scheduler.PushTask(l.ctx, schedTask); err != nil {
		l.Logger.Errorf("push task to queue failed: %v", err)
		return &types.BaseResp{Code: 500, Msg: "任务入队失败"}, nil
	}

	// 保存任务信息到 Redis
	taskInfoKey := "cscan:task:info:" + task.TaskId
	taskInfoData, _ := json.Marshal(map[string]string{
		"workspaceId": workspaceId,
		"mainTaskId":  task.Id.Hex(),
	})
	l.svcCtx.RedisClient.Set(l.ctx, taskInfoKey, taskInfoData, 24*time.Hour)

	return &types.BaseResp{Code: 0, Msg: "任务已启动"}, nil
}

// MainTaskPauseLogic 暂停任务
type MainTaskPauseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMainTaskPauseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MainTaskPauseLogic {
	return &MainTaskPauseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MainTaskPauseLogic) MainTaskPause(req *types.MainTaskControlReq, workspaceId string) (resp *types.BaseResp, err error) {
	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)

	// 获取任务
	task, err := taskModel.FindById(l.ctx, req.Id)
	if err != nil {
		return &types.BaseResp{Code: 400, Msg: "任务不存在"}, nil
	}

	// 检查状态：只有STARTED状态可以暂停
	if task.Status != model.TaskStatusStarted {
		return &types.BaseResp{Code: 400, Msg: "只有执行中的任务可以暂停"}, nil
	}

	// 发送暂停信号到Redis
	ctrlKey := "cscan:task:ctrl:" + task.TaskId
	l.svcCtx.RedisClient.Set(l.ctx, ctrlKey, "PAUSE", 24*time.Hour)

	// 更新状态为PAUSED
	update := bson.M{"status": model.TaskStatusPaused}
	if err := taskModel.Update(l.ctx, req.Id, update); err != nil {
		return &types.BaseResp{Code: 500, Msg: "更新任务状态失败"}, nil
	}

	l.Logger.Infof("Task paused: taskId=%s", task.TaskId)
	return &types.BaseResp{Code: 0, Msg: "任务已暂停"}, nil
}

// MainTaskResumeLogic 继续任务
type MainTaskResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMainTaskResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MainTaskResumeLogic {
	return &MainTaskResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MainTaskResumeLogic) MainTaskResume(req *types.MainTaskControlReq, workspaceId string) (resp *types.BaseResp, err error) {
	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)

	// 获取任务
	task, err := taskModel.FindById(l.ctx, req.Id)
	if err != nil {
		return &types.BaseResp{Code: 400, Msg: "任务不存在"}, nil
	}

	// 检查状态：只有PAUSED状态可以继续
	if task.Status != model.TaskStatusPaused {
		return &types.BaseResp{Code: 400, Msg: "只有已暂停的任务可以继续"}, nil
	}

	// 清除暂停信号
	ctrlKey := "cscan:task:ctrl:" + task.TaskId
	l.svcCtx.RedisClient.Del(l.ctx, ctrlKey)

	// 更新状态为PENDING
	update := bson.M{"status": model.TaskStatusPending}
	if err := taskModel.Update(l.ctx, req.Id, update); err != nil {
		return &types.BaseResp{Code: 500, Msg: "更新任务状态失败"}, nil
	}

	// 重新发送任务到队列（带上已保存的状态）
	config := task.Config
	if task.TaskState != "" {
		// 将已保存的状态注入到配置中
		var configMap map[string]interface{}
		if json.Unmarshal([]byte(config), &configMap) == nil {
			configMap["resumeState"] = task.TaskState
			if newConfig, err := json.Marshal(configMap); err == nil {
				config = string(newConfig)
			}
		}
	}

	schedTask := &scheduler.TaskInfo{
		TaskId:      task.TaskId,
		MainTaskId:  task.Id.Hex(),
		WorkspaceId: workspaceId,
		TaskName:    task.Name,
		Config:      config,
		Priority:    1,
	}
	l.Logger.Infof("Resuming task: taskId=%s, workspaceId=%s, hasState=%v", task.TaskId, workspaceId, task.TaskState != "")
	if err := l.svcCtx.Scheduler.PushTask(l.ctx, schedTask); err != nil {
		l.Logger.Errorf("push task to queue failed: %v", err)
		return &types.BaseResp{Code: 500, Msg: "任务入队失败"}, nil
	}

	return &types.BaseResp{Code: 0, Msg: "任务已继续"}, nil
}

// MainTaskStopLogic 停止任务
type MainTaskStopLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMainTaskStopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MainTaskStopLogic {
	return &MainTaskStopLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MainTaskStopLogic) MainTaskStop(req *types.MainTaskControlReq, workspaceId string) (resp *types.BaseResp, err error) {
	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)

	// 获取任务
	task, err := taskModel.FindById(l.ctx, req.Id)
	if err != nil {
		return &types.BaseResp{Code: 400, Msg: "任务不存在"}, nil
	}

	// 检查状态：STARTED, PAUSED, PENDING 状态可以停止
	if task.Status != model.TaskStatusStarted && task.Status != model.TaskStatusPaused && task.Status != model.TaskStatusPending {
		return &types.BaseResp{Code: 400, Msg: "当前状态不可停止"}, nil
	}

	// 发送停止信号到Redis
	ctrlKey := "cscan:task:ctrl:" + task.TaskId
	l.svcCtx.RedisClient.Set(l.ctx, ctrlKey, "STOP", 24*time.Hour)

	// 更新状态为STOPPED
	update := bson.M{
		"status": model.TaskStatusStopped,
		"result": "任务已手动停止",
	}
	if err := taskModel.Update(l.ctx, req.Id, update); err != nil {
		return &types.BaseResp{Code: 500, Msg: "更新任务状态失败"}, nil
	}

	l.Logger.Infof("Task stopped: taskId=%s", task.TaskId)
	return &types.BaseResp{Code: 0, Msg: "任务已停止"}, nil
}




// TaskStatLogic 任务统计逻辑
type TaskStatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTaskStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaskStatLogic {
	return &TaskStatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TaskStatLogic) TaskStat(workspaceId string) (resp *types.TaskStatResp, err error) {
	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)

	// 统计总数
	total, _ := taskModel.Count(l.ctx, bson.M{})

	// 按状态统计
	completed, _ := taskModel.Count(l.ctx, bson.M{"status": model.TaskStatusSuccess})
	running, _ := taskModel.Count(l.ctx, bson.M{"status": model.TaskStatusStarted})
	failed, _ := taskModel.Count(l.ctx, bson.M{"status": model.TaskStatusFailure})
	pending, _ := taskModel.Count(l.ctx, bson.M{"status": bson.M{"$in": []string{model.TaskStatusPending, model.TaskStatusCreated}}})

	// 近7天每日趋势统计
	now := time.Now()
	trendDays := make([]string, 7)
	trendCompleted := make([]int, 7)
	trendFailed := make([]int, 7)

	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
		dayEnd := dayStart.AddDate(0, 0, 1)

		idx := 6 - i
		trendDays[idx] = dayStart.Format("01-02")

		// 统计当天完成的任务
		completedCount, _ := taskModel.Count(l.ctx, bson.M{
			"status":      model.TaskStatusSuccess,
			"update_time": bson.M{"$gte": dayStart, "$lt": dayEnd},
		})
		trendCompleted[idx] = int(completedCount)

		// 统计当天失败的任务
		failedCount, _ := taskModel.Count(l.ctx, bson.M{
			"status":      model.TaskStatusFailure,
			"update_time": bson.M{"$gte": dayStart, "$lt": dayEnd},
		})
		trendFailed[idx] = int(failedCount)
	}

	return &types.TaskStatResp{
		Code:           0,
		Msg:            "success",
		Total:          int(total),
		Completed:      int(completed),
		Running:        int(running),
		Failed:         int(failed),
		Pending:        int(pending),
		TrendDays:      trendDays,
		TrendCompleted: trendCompleted,
		TrendFailed:    trendFailed,
	}, nil
}

// MainTaskUpdateLogic 更新任务逻辑 
type MainTaskUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMainTaskUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MainTaskUpdateLogic {
	return &MainTaskUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MainTaskUpdateLogic) MainTaskUpdate(req *types.MainTaskUpdateReq, workspaceId string) (resp *types.BaseResp, err error) {
	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)

	// 获取任务
	task, err := taskModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Logger.Errorf("MainTaskUpdate: task not found, id=%s, error=%v", req.Id, err)
		return &types.BaseResp{Code: 40001, Msg: "任务不存在"}, nil
	}

	// 检查状态：只有CREATED状态可以编辑 
	if task.Status != model.TaskStatusCreated {
		l.Logger.Infof("MainTaskUpdate: task status not allowed, id=%s, status=%s", req.Id, task.Status)
		return &types.BaseResp{Code: 40002, Msg: "任务状态不允许编辑，只有待启动状态的任务可以编辑"}, nil
	}

	// 构建更新字段
	update := bson.M{}

	if req.Name != "" {
		update["name"] = req.Name
	}

	if req.Target != "" {
		update["target"] = req.Target
	}

	if req.ProfileId != "" {
		// 验证配置是否存在
		profile, err := l.svcCtx.ProfileModel.FindById(l.ctx, req.ProfileId)
		if err != nil {
			l.Logger.Errorf("MainTaskUpdate: profile not found, profileId=%s, error=%v", req.ProfileId, err)
			return &types.BaseResp{Code: 400, Msg: "任务配置不存在"}, nil
		}
		update["profile_id"] = req.ProfileId
		update["profile_name"] = profile.Name

		// 更新任务配置
		taskConfig := map[string]interface{}{
			"target": task.Target,
		}
		if req.Target != "" {
			taskConfig["target"] = req.Target
		}
		// 合并 profile 配置
		if profile.Config != "" {
			var profileConfig map[string]interface{}
			if err := json.Unmarshal([]byte(profile.Config), &profileConfig); err == nil {
				for k, v := range profileConfig {
					taskConfig[k] = v
				}
			}
		}
		// 注入自定义POC和标签映射
		taskConfig = common.InjectPocConfig(l.ctx, l.svcCtx, taskConfig, l.Logger)
		configBytes, _ := json.Marshal(taskConfig)
		update["config"] = string(configBytes)
	} else if req.Target != "" {
		// 只更新了target，需要重新生成config
		taskConfig := map[string]interface{}{
			"target": req.Target,
		}
		// 获取当前profile配置
		if task.ProfileId != "" {
			profile, err := l.svcCtx.ProfileModel.FindById(l.ctx, task.ProfileId)
			if err == nil && profile.Config != "" {
				var profileConfig map[string]interface{}
				if err := json.Unmarshal([]byte(profile.Config), &profileConfig); err == nil {
					for k, v := range profileConfig {
						taskConfig[k] = v
					}
				}
			}
		}
		// 注入自定义POC和标签映射
		taskConfig = common.InjectPocConfig(l.ctx, l.svcCtx, taskConfig, l.Logger)
		configBytes, _ := json.Marshal(taskConfig)
		update["config"] = string(configBytes)
	}

	if len(update) == 0 {
		return &types.BaseResp{Code: 400, Msg: "没有需要更新的字段"}, nil
	}

	// 再次检查状态（防止并发修改）
	task, err = taskModel.FindById(l.ctx, req.Id)
	if err != nil {
		return &types.BaseResp{Code: 40001, Msg: "任务不存在"}, nil
	}
	if task.Status != model.TaskStatusCreated {
		return &types.BaseResp{Code: 40002, Msg: "任务状态已变更，无法编辑"}, nil
	}

	// 执行更新
	if err := taskModel.Update(l.ctx, req.Id, update); err != nil {
		l.Logger.Errorf("MainTaskUpdate: update failed, id=%s, error=%v", req.Id, err)
		return &types.BaseResp{Code: 500, Msg: "更新任务失败"}, nil
	}

	l.Logger.Infof("MainTaskUpdate: task updated, id=%s, workspaceId=%s", req.Id, workspaceId)
	return &types.BaseResp{Code: 0, Msg: "任务更新成功"}, nil
}

// GetTaskLogsLogic 获取任务日志逻辑 
type GetTaskLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTaskLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskLogsLogic {
	return &GetTaskLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTaskLogsLogic) GetTaskLogs(req *types.GetTaskLogsReq) (resp *types.GetTaskLogsResp, err error) {
	if req.TaskId == "" {
		return &types.GetTaskLogsResp{Code: 400, Msg: "任务ID不能为空", List: []types.TaskLogEntry{}}, nil
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	// 从Redis Stream读取任务专属日志 (cscan:task:logs:{taskId})
	streamKey := "cscan:task:logs:" + req.TaskId
	logs, err := l.svcCtx.RedisClient.XRevRange(l.ctx, streamKey, "+", "-").Result()
	if err != nil {
		l.Logger.Errorf("GetTaskLogs: failed to read logs from Redis, taskId=%s, error=%v", req.TaskId, err)
		// 返回空列表而不是错误
		return &types.GetTaskLogsResp{Code: 0, Msg: "success", List: []types.TaskLogEntry{}}, nil
	}

	// 解析日志条目
	result := make([]types.TaskLogEntry, 0)
	count := limit
	if len(logs) < count {
		count = len(logs)
	}

	// XRevRange返回的是倒序，我们需要正序显示，所以从后往前遍历
	for i := count - 1; i >= 0; i-- {
		if data, ok := logs[i].Values["data"].(string); ok {
			var entry types.TaskLogEntry
			if err := json.Unmarshal([]byte(data), &entry); err == nil {
				// 只返回匹配taskId的日志
				if entry.TaskId == req.TaskId {
					result = append(result, entry)
				}
			}
		}
	}

	l.Logger.Infof("GetTaskLogs: returned %d logs for taskId=%s", len(result), req.TaskId)
	return &types.GetTaskLogsResp{Code: 0, Msg: "success", List: result}, nil
}
