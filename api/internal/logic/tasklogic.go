package logic

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

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
	taskConfig = l.injectPocConfig(taskConfig)
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
	taskConfig = l.injectPocConfig(taskConfig)
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

// injectPocConfig 注入POC模板ID到任务配置（不存储完整内容，避免文档过大）
func (l *MainTaskCreateLogic) injectPocConfig(taskConfig map[string]interface{}) map[string]interface{} {
	pocscan, ok := taskConfig["pocscan"].(map[string]interface{})
	if !ok || pocscan == nil {
		return taskConfig
	}

	// 检查是否启用POC扫描和Nuclei
	enable, _ := pocscan["enable"].(bool)
	useNuclei, _ := pocscan["useNuclei"].(bool)
	if !enable || !useNuclei {
		return taskConfig
	}

	// 检查是否启用自动扫描模式
	autoScan, _ := pocscan["autoScan"].(bool)
	automaticScan, _ := pocscan["automaticScan"].(bool)

	// 如果启用了自动扫描，不预先注入模板ID，让Worker根据资产指纹动态获取
	if autoScan || automaticScan {
		l.Logger.Infof("Auto-scan enabled (autoScan=%v, automaticScan=%v), skipping template ID injection", autoScan, automaticScan)
		
		// 只注入标签映射（用于自定义标签映射模式）
		if autoScan {
			tagMappings, err := l.svcCtx.TagMappingModel.FindEnabled(l.ctx)
			if err == nil && len(tagMappings) > 0 {
				mappings := make(map[string][]string)
				for _, tm := range tagMappings {
					mappings[tm.AppName] = tm.NucleiTags
				}
				pocscan["tagMappings"] = mappings
				l.Logger.Infof("Injected %d tag mappings for auto-scan", len(mappings))
			}
		}
		
		taskConfig["pocscan"] = pocscan
		return taskConfig
	}

	customPocOnly, _ := pocscan["customPocOnly"].(bool)
	var nucleiTemplateIds []string
	var customPocIds []string

	if customPocOnly {
		// 只使用自定义POC - 存储ID列表
		customPocs, err := l.svcCtx.CustomPocModel.FindEnabled(l.ctx)
		if err == nil && len(customPocs) > 0 {
			for _, poc := range customPocs {
				customPocIds = append(customPocIds, poc.Id.Hex())
			}
			l.Logger.Infof("Injected %d custom POC IDs (CustomPocOnly mode)", len(customPocIds))
		}
	} else {
		// 从数据库获取默认模板ID（根据严重级别筛选）
		severityStr, _ := pocscan["severity"].(string)
		if severityStr != "" {
			severities := strings.Split(severityStr, ",")
			nucleiTemplates, err := l.svcCtx.NucleiTemplateModel.FindBySeverity(l.ctx, severities)
			if err == nil && len(nucleiTemplates) > 0 {
				for _, t := range nucleiTemplates {
					nucleiTemplateIds = append(nucleiTemplateIds, t.Id.Hex())
				}
				l.Logger.Infof("Injected %d nuclei template IDs (severity: %s)", len(nucleiTemplateIds), severityStr)
			}
		}

		// 添加自定义POC ID
		customPocs, err := l.svcCtx.CustomPocModel.FindEnabled(l.ctx)
		if err == nil && len(customPocs) > 0 {
			for _, poc := range customPocs {
				customPocIds = append(customPocIds, poc.Id.Hex())
			}
			l.Logger.Infof("Added %d custom POC IDs", len(customPocIds))
		}
	}

	// 存储ID列表而不是完整内容
	if len(nucleiTemplateIds) > 0 {
		pocscan["nucleiTemplateIds"] = nucleiTemplateIds
	}
	if len(customPocIds) > 0 {
		pocscan["customPocIds"] = customPocIds
	}

	taskConfig["pocscan"] = pocscan
	return taskConfig
}

// injectPocConfig 注入POC模板ID到任务配置 (MainTaskRetryLogic)
func (l *MainTaskRetryLogic) injectPocConfig(taskConfig map[string]interface{}) map[string]interface{} {
	pocscan, ok := taskConfig["pocscan"].(map[string]interface{})
	if !ok || pocscan == nil {
		return taskConfig
	}

	// 检查是否启用POC扫描和Nuclei
	enable, _ := pocscan["enable"].(bool)
	useNuclei, _ := pocscan["useNuclei"].(bool)
	if !enable || !useNuclei {
		return taskConfig
	}

	// 检查是否启用自动扫描模式
	autoScan, _ := pocscan["autoScan"].(bool)
	automaticScan, _ := pocscan["automaticScan"].(bool)

	// 如果启用了自动扫描，不预先注入模板ID，让Worker根据资产指纹动态获取
	if autoScan || automaticScan {
		l.Logger.Infof("Auto-scan enabled (autoScan=%v, automaticScan=%v), skipping template ID injection", autoScan, automaticScan)
		
		// 只注入标签映射（用于自定义标签映射模式）
		if autoScan {
			tagMappings, err := l.svcCtx.TagMappingModel.FindEnabled(l.ctx)
			if err == nil && len(tagMappings) > 0 {
				mappings := make(map[string][]string)
				for _, tm := range tagMappings {
					mappings[tm.AppName] = tm.NucleiTags
				}
				pocscan["tagMappings"] = mappings
				l.Logger.Infof("Injected %d tag mappings for auto-scan", len(mappings))
			}
		}
		
		taskConfig["pocscan"] = pocscan
		return taskConfig
	}

	customPocOnly, _ := pocscan["customPocOnly"].(bool)
	var nucleiTemplateIds []string
	var customPocIds []string

	if customPocOnly {
		// 只使用自定义POC - 存储ID列表
		customPocs, err := l.svcCtx.CustomPocModel.FindEnabled(l.ctx)
		if err == nil && len(customPocs) > 0 {
			for _, poc := range customPocs {
				customPocIds = append(customPocIds, poc.Id.Hex())
			}
			l.Logger.Infof("Injected %d custom POC IDs (CustomPocOnly mode)", len(customPocIds))
		}
	} else {
		// 从数据库获取默认模板ID（根据严重级别筛选）
		severityStr, _ := pocscan["severity"].(string)
		if severityStr != "" {
			severities := strings.Split(severityStr, ",")
			nucleiTemplates, err := l.svcCtx.NucleiTemplateModel.FindBySeverity(l.ctx, severities)
			if err == nil && len(nucleiTemplates) > 0 {
				for _, t := range nucleiTemplates {
					nucleiTemplateIds = append(nucleiTemplateIds, t.Id.Hex())
				}
				l.Logger.Infof("Injected %d nuclei template IDs (severity: %s)", len(nucleiTemplateIds), severityStr)
			}
		}

		// 添加自定义POC ID
		customPocs, err := l.svcCtx.CustomPocModel.FindEnabled(l.ctx)
		if err == nil && len(customPocs) > 0 {
			for _, poc := range customPocs {
				customPocIds = append(customPocIds, poc.Id.Hex())
			}
			l.Logger.Infof("Added %d custom POC IDs", len(customPocIds))
		}
	}

	// 存储ID列表而不是完整内容
	if len(nucleiTemplateIds) > 0 {
		pocscan["nucleiTemplateIds"] = nucleiTemplateIds
	}
	if len(customPocIds) > 0 {
		pocscan["customPocIds"] = customPocIds
	}

	taskConfig["pocscan"] = pocscan
	return taskConfig
}
