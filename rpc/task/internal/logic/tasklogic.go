package logic

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"cscan/model"
	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type TaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaskLogic {
	return &TaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TaskLogic) CheckTask(in *pb.CheckTaskReq) (*pb.CheckTaskResp, error) {
	// 从 Redis 队列获取任务
	queueKey := "cscan:task:queue"
	
	// 获取优先级最高的任务
	results, err := l.svcCtx.RedisClient.ZPopMin(l.ctx, queueKey, 1).Result()
	if err != nil || len(results) == 0 {
		// 没有任务
		return &pb.CheckTaskResp{
			IsExist:    false,
			IsFinished: true,
		}, nil
	}

	// 解析任务信息
	var taskInfo struct {
		TaskId      string `json:"taskId"`
		MainTaskId  string `json:"mainTaskId"`
		WorkspaceId string `json:"workspaceId"`
		TaskName    string `json:"taskName"`
		Config      string `json:"config"`
	}
	if err := json.Unmarshal([]byte(results[0].Member.(string)), &taskInfo); err != nil {
		return &pb.CheckTaskResp{IsExist: false, IsFinished: true}, nil
	}

	l.Logger.Infof("Dispatching task: taskId=%s, workspaceId=%s, config=%s", 
		taskInfo.TaskId, taskInfo.WorkspaceId, taskInfo.Config)

	return &pb.CheckTaskResp{
		IsExist:     true,
		IsFinished:  false,
		TaskId:      taskInfo.TaskId,
		State:       "PENDING",
		WorkspaceId: taskInfo.WorkspaceId,
		Config:      taskInfo.Config,
	}, nil
}

func (l *TaskLogic) UpdateTask(in *pb.UpdateTaskReq) (*pb.UpdateTaskResp, error) {
	l.Logger.Infof("UpdateTask: taskId=%s, state=%s, worker=%s, result=%s", in.TaskId, in.State, in.Worker, in.Result)
	
	// 从 Redis 获取任务的 workspaceId（任务创建时保存）
	taskKey := fmt.Sprintf("cscan:task:info:%s", in.TaskId)
	taskData, err := l.svcCtx.RedisClient.Get(l.ctx, taskKey).Result()
	
	if err == nil && taskData != "" {
		var taskInfo struct {
			WorkspaceId string `json:"workspaceId"`
			MainTaskId  string `json:"mainTaskId"`
		}
		if json.Unmarshal([]byte(taskData), &taskInfo) == nil && taskInfo.WorkspaceId != "" {
			// 更新 MongoDB 中的主任务状态
			taskModel := l.svcCtx.GetMainTaskModel(taskInfo.WorkspaceId)
			
			// 计算进度
			progress := 0
			switch in.State {
			case "STARTED":
				progress = 10
			case "PAUSED":
				progress = 50 // 暂停时保持中间进度
			case "SUCCESS":
				progress = 100
			case "FAILURE":
				progress = 100
			case "STOPPED":
				progress = 100
			}
			
			update := bson.M{
				"status":   in.State,
				"progress": progress,
				"worker":   in.Worker,
			}
			
			// 如果是暂停状态，保存任务状态到task_state字段
			if in.State == "PAUSED" {
				update["task_state"] = in.Result // Result字段用于传递状态JSON
			} else {
				update["result"] = in.Result
			}
			
			if err := taskModel.UpdateByTaskId(l.ctx, in.TaskId, update); err != nil {
				l.Logger.Errorf("Update task in MongoDB failed: %v", err)
			}
		}
	}
	
	return &pb.UpdateTaskResp{
		Success: true,
		Message: "ok",
	}, nil
}

func (l *TaskLogic) NewTask(in *pb.NewTaskReq) (*pb.NewTaskResp, error) {
	if in.WorkspaceId == "" {
		return &pb.NewTaskResp{Success: false, Message: "workspaceId is required"}, nil
	}

	taskModel := l.svcCtx.GetExecutorTaskModel(in.WorkspaceId)
	
	task := &model.ExecutorTask{
		TaskId:     in.TaskId,
		MainTaskId: in.MainTaskId,
		TaskName:   in.TaskName,
		Config:     in.Config,
	}

	if err := taskModel.Insert(l.ctx, task); err != nil {
		return &pb.NewTaskResp{Success: false, Message: err.Error()}, nil
	}

	return &pb.NewTaskResp{Success: true, Message: "ok"}, nil
}

func (l *TaskLogic) SaveTaskResult(in *pb.SaveTaskResultReq) (*pb.SaveTaskResultResp, error) {
	l.Logger.Infof("SaveTaskResult: workspaceId=%s, mainTaskId=%s, assets=%d", in.WorkspaceId, in.MainTaskId, len(in.Assets))
	
	if in.WorkspaceId == "" {
		l.Logger.Error("SaveTaskResult: workspaceId is empty")
		return &pb.SaveTaskResultResp{Success: false, Message: "workspaceId is required"}, nil
	}

	assetModel := l.svcCtx.GetAssetModel(in.WorkspaceId)
	assetHistoryModel := l.svcCtx.GetAssetHistoryModel(in.WorkspaceId)
	
	var newCount, updateCount int
	for _, asset := range in.Assets {
		doc := convertPbAssetToModel(asset, in.MainTaskId)
		
		// 根据host:port查找是否存在（同IP同端口覆盖）
		existing, err := assetModel.FindByHostPort(l.ctx, doc.Host, doc.Port)
		if err != nil || existing == nil {
			// 新增
			if err := assetModel.Insert(l.ctx, doc); err == nil {
				newCount++
			}
		} else {
			// 保存历史记录
			history := &model.AssetHistory{
				AssetId:    existing.Id.Hex(),
				Authority:  existing.Authority,
				Host:       existing.Host,
				Port:       existing.Port,
				Service:    existing.Service,
				Title:      existing.Title,
				App:        existing.App,
				HttpStatus: existing.HttpStatus,
				HttpHeader: existing.HttpHeader,
				HttpBody:   existing.HttpBody,
				Banner:     existing.Banner,
				IconHash:   existing.IconHash,
				Screenshot: existing.Screenshot,
				TaskId:     existing.TaskId,
			}
			assetHistoryModel.Insert(l.ctx, history)
			
			// 更新资产，设置update=true
			update := buildAssetUpdate(doc)
			if err := assetModel.Update(l.ctx, existing.Id.Hex(), update); err == nil {
				updateCount++
			}
		}
	}

	return &pb.SaveTaskResultResp{
		Success:     true,
		Message:     fmt.Sprintf("资产: %d (新增%d, 更新%d)", len(in.Assets), newCount, updateCount),
		TotalAsset:  int32(len(in.Assets)),
		NewAsset:    int32(newCount),
		UpdateAsset: int32(updateCount),
	}, nil
}

func (l *TaskLogic) SaveVulResult(in *pb.SaveVulResultReq) (*pb.SaveVulResultResp, error) {
	if in.WorkspaceId == "" {
		return &pb.SaveVulResultResp{Success: false, Message: "workspaceId is required"}, nil
	}

	vulModel := l.svcCtx.GetVulModel(in.WorkspaceId)
	
	var count int
	for _, vul := range in.Vuls {
		doc := &model.Vul{
			Authority: vul.Authority,
			Host:      vul.Host,
			Port:      int(vul.Port),
			Url:       vul.Url,
			PocFile:   vul.PocFile,
			Source:    vul.Source,
			Severity:  vul.Severity,
			Extra:     vul.Extra,
			Result:    vul.Result,
			TaskId:    vul.TaskId,
		}
		// 使用 Upsert 去重插入（基于 host+port+pocFile+url）
		if err := vulModel.Upsert(l.ctx, doc); err == nil {
			count++
		}
	}

	return &pb.SaveVulResultResp{
		Success: true,
		Message: fmt.Sprintf("漏洞: %d", count),
		Total:   int32(count),
	}, nil
}

func (l *TaskLogic) KeepAlive(in *pb.KeepAliveReq) (*pb.KeepAliveResp, error) {
	if in.WorkerName == "" {
		return &pb.KeepAliveResp{Status: "error"}, nil
	}

	// 保存Worker状态到Redis
	key := fmt.Sprintf("worker:%s", in.WorkerName)
	status := map[string]interface{}{
		"workerName":         in.WorkerName,
		"cpuLoad":            in.CpuLoad,
		"memUsed":            in.MemUsed,
		"taskStartedNumber":  in.TaskStartedNumber,
		"taskExecutedNumber": in.TaskExecutedNumber,
		"isDaemon":           in.IsDaemon,
		"updateTime":         time.Now().Format("2006-01-02 15:04:05"),
	}

	data, _ := json.Marshal(status)
	l.svcCtx.RedisClient.Set(l.ctx, key, data, 10*time.Minute)

	// 检查是否有控制指令
	ctrlKey := fmt.Sprintf("worker_ctrl:%s", in.WorkerName)
	ctrlData, err := l.svcCtx.RedisClient.Get(l.ctx, ctrlKey).Result()
	
	resp := &pb.KeepAliveResp{Status: "ok"}
	if err == nil && ctrlData != "" {
		var ctrl map[string]bool
		if json.Unmarshal([]byte(ctrlData), &ctrl) == nil {
			resp.ManualStopFlag = ctrl["stop"]
			resp.ManualReloadFlag = ctrl["reload"]
			resp.ManualInitEnvFlag = ctrl["init"]
			resp.ManualSyncFlag = ctrl["sync"]
		}
		// 清除控制指令
		l.svcCtx.RedisClient.Del(l.ctx, ctrlKey)
	}

	return resp, nil
}

func (l *TaskLogic) GetWorkerConfig(in *pb.GetWorkerConfigReq) (*pb.GetWorkerConfigResp, error) {
	// 返回Worker配置
	config := map[string]interface{}{
		"concurrency": 10,
		"timeout":     3600,
	}
	data, _ := json.Marshal(config)
	return &pb.GetWorkerConfigResp{Config: string(data)}, nil
}

func (l *TaskLogic) RequestResource(in *pb.RequestResourceReq) (*pb.RequestResourceResp, error) {
	// 资源文件请求，用于分发POC文件等
	return &pb.RequestResourceResp{
		Path: in.Name,
		Hash: "",
		Data: nil,
	}, nil
}

func convertPbAssetToModel(pb *pb.AssetDocument, taskId string) *model.Asset {
	doc := &model.Asset{
		Authority:  pb.Authority,
		Host:       pb.Host,
		Port:       int(pb.Port),
		Category:   pb.Category,
		Service:    pb.Service,
		Server:     pb.Server,
		Banner:     pb.Banner,
		Title:      pb.Title,
		App:        pb.App,
		HttpStatus: pb.HttpStatus,
		HttpHeader: pb.HttpHeader,
		HttpBody:   pb.HttpBody,
		Cert:       pb.Cert,
		IconHash:   pb.IconHash,
		Screenshot: pb.Screenshot,
		IsCDN:      pb.IsCdn,
		CName:      pb.Cname,
		IsCloud:    pb.IsCloud,
		IsHTTP:     pb.IsHttp,
		TaskId:     taskId,
	}

	for _, ip := range pb.Ipv4 {
		doc.Ip.IpV4 = append(doc.Ip.IpV4, model.IPV4{
			IPName:   ip.Ip,
			IPInt:    ip.IpInt,
			Location: ip.Location,
		})
	}

	for _, ip := range pb.Ipv6 {
		doc.Ip.IpV6 = append(doc.Ip.IpV6, model.IPV6{
			IPName:   ip.Ip,
			Location: ip.Location,
		})
	}

	return doc
}

func buildAssetUpdate(doc *model.Asset) bson.M {
	update := bson.M{
		"update_time": time.Now(),
		"new":         false,
		"update":      true,
		"is_http":     doc.IsHTTP,
	}
	if doc.Service != "" {
		update["service"] = doc.Service
	}
	if doc.Title != "" {
		update["title"] = doc.Title
	}
	if len(doc.App) > 0 {
		update["app"] = doc.App
	}
	if doc.HttpStatus != "" {
		update["status"] = doc.HttpStatus
	}
	if doc.HttpHeader != "" {
		update["header"] = doc.HttpHeader
	}
	if doc.HttpBody != "" {
		update["body"] = doc.HttpBody
	}
	if doc.Banner != "" {
		update["banner"] = doc.Banner
	}
	if doc.IconHash != "" {
		update["icon_hash"] = doc.IconHash
	}
	if doc.Screenshot != "" {
		update["screenshot"] = doc.Screenshot
	}
	if doc.Server != "" {
		update["server"] = doc.Server
	}
	return update
}

// GetTemplatesByTags 根据标签获取模板
func (l *TaskLogic) GetTemplatesByTags(in *pb.GetTemplatesByTagsReq) (*pb.GetTemplatesByTagsResp, error) {
	l.Logger.Infof("GetTemplatesByTags: tags=%v, severities=%v", in.Tags, in.Severities)

	var templates []string

	// 根据标签和严重级别从数据库获取模板
	filter := bson.M{"enabled": true}

	// 如果有标签条件
	if len(in.Tags) > 0 {
		filter["tags"] = bson.M{"$in": in.Tags}
	}

	// 如果有严重级别条件
	if len(in.Severities) > 0 {
		filter["severity"] = bson.M{"$in": in.Severities}
	}

	// 查询模板
	nucleiTemplates, err := l.svcCtx.NucleiTemplateModel.FindEnabledByFilter(l.ctx, filter)
	if err != nil {
		l.Logger.Errorf("Failed to get templates by tags: %v", err)
		return &pb.GetTemplatesByTagsResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	for _, t := range nucleiTemplates {
		if t.Content != "" {
			templates = append(templates, t.Content)
		}
	}

	l.Logger.Infof("Found %d templates matching tags=%v, severities=%v", len(templates), in.Tags, in.Severities)

	return &pb.GetTemplatesByTagsResp{
		Success:   true,
		Message:   "success",
		Templates: templates,
		Count:     int32(len(templates)),
	}, nil
}

// GetCustomFingerprints 获取指纹（包括内置和自定义）
func (l *TaskLogic) GetCustomFingerprints(in *pb.GetCustomFingerprintsReq) (*pb.GetCustomFingerprintsResp, error) {
	l.Logger.Infof("GetCustomFingerprints: enabledOnly=%v", in.EnabledOnly)

	// 查询所有指纹（包括内置和自定义）
	filter := bson.M{}
	if in.EnabledOnly {
		filter["enabled"] = true
	}

	fingerprints, err := l.svcCtx.FingerprintModel.Find(l.ctx, filter, 0, 0)
	if err != nil {
		l.Logger.Errorf("Failed to get fingerprints: %v", err)
		return &pb.GetCustomFingerprintsResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	var pbFingerprints []*pb.FingerprintDocument
	for _, fp := range fingerprints {
		pbFp := &pb.FingerprintDocument{
			Id:        fp.Id.Hex(),
			Name:      fp.Name,
			Category:  fp.Category,
			Rule:      fp.Rule,
			Source:    fp.Source,
			Headers:   fp.Headers,
			Cookies:   fp.Cookies,
			Html:      fp.HTML,
			Scripts:   fp.Scripts,
			ScriptSrc: fp.ScriptSrc,
			Meta:      fp.Meta,
			Css:       fp.CSS,
			Url:       fp.URL,
			IsBuiltin: fp.IsBuiltin,
			Enabled:   fp.Enabled,
		}
		pbFingerprints = append(pbFingerprints, pbFp)
	}

	l.Logger.Infof("Found %d fingerprints (builtin + custom)", len(pbFingerprints))

	return &pb.GetCustomFingerprintsResp{
		Success:      true,
		Message:      "success",
		Fingerprints: pbFingerprints,
		Count:        int32(len(pbFingerprints)),
	}, nil
}

// ValidateFingerprint 验证指纹匹配
func (l *TaskLogic) ValidateFingerprint(in *pb.ValidateFingerprintReq) (*pb.ValidateFingerprintResp, error) {
	l.Logger.Infof("ValidateFingerprint: url=%s, fingerprintId=%s, scope=%s", in.Url, in.FingerprintId, in.Scope)

	if in.Url == "" {
		return &pb.ValidateFingerprintResp{
			Success: false,
			Message: "URL不能为空",
		}, nil
	}

	startTime := time.Now()

	// 请求目标URL获取响应数据
	data, err := l.fetchFingerprintData(in.Url)
	if err != nil {
		return &pb.ValidateFingerprintResp{
			Success: false,
			Message: "请求目标失败: " + err.Error(),
		}, nil
	}

	if in.FingerprintId != "" {
		// 单个指纹验证
		return l.validateSingleFingerprint(in.FingerprintId, in.Url, data)
	} else {
		// 批量验证所有指纹
		return l.validateAllFingerprints(in.Url, in.Scope, data, startTime)
	}
}

// validateSingleFingerprint 验证单个指纹
func (l *TaskLogic) validateSingleFingerprint(fingerprintId, url string, data *FingerprintData) (*pb.ValidateFingerprintResp, error) {
	// 获取指纹
	fp, err := l.svcCtx.FingerprintModel.FindById(l.ctx, fingerprintId)
	if err != nil {
		return &pb.ValidateFingerprintResp{
			Success: false,
			Message: "指纹不存在",
		}, nil
	}

	// 使用指纹引擎进行匹配
	engine := l.newSingleFingerprintEngine(fp)
	matched, matchedConditions := engine.MatchWithDetails(data)

	// 构建详细的调试信息
	var details strings.Builder
	details.WriteString("═══════════════════════════════════════════════════════════\n")
	details.WriteString(fmt.Sprintf("目标URL: %s\n", url))
	details.WriteString(fmt.Sprintf("指纹名称: %s\n", fp.Name))
	details.WriteString(fmt.Sprintf("匹配规则: %s\n", fp.Rule))
	details.WriteString("═══════════════════════════════════════════════════════════\n\n")

	if matched {
		details.WriteString("【匹配结果】✓ 匹配成功\n\n")
		// 显示匹配的条件
		details.WriteString("───────────────────────────────────────────────────────────\n")
		details.WriteString("【命中条件】 ★★★\n")
		for _, cond := range matchedConditions {
			details.WriteString(fmt.Sprintf("  ✓ %s\n", cond))
		}
		details.WriteString("\n")
	} else {
		details.WriteString("【匹配结果】✗ 未匹配\n\n")
	}

	// 响应基本信息
	details.WriteString("───────────────────────────────────────────────────────────\n")
	details.WriteString("【响应基本信息】\n")
	details.WriteString(fmt.Sprintf("  Title: %s\n", data.Title))
	details.WriteString(fmt.Sprintf("  Server: %s\n", data.Server))
	details.WriteString(fmt.Sprintf("  Body长度: %d 字节\n", len(data.Body)))
	details.WriteString(fmt.Sprintf("  Icon Hash (MMH3): %s\n", data.FaviconHash))
	details.WriteString(fmt.Sprintf("  Cookies: %s\n", l.truncateString(data.Cookies, 200)))

	// 完整Header
	details.WriteString("\n───────────────────────────────────────────────────────────\n")
	details.WriteString("【响应头 Headers】\n")
	if data.HeaderString != "" {
		details.WriteString(data.HeaderString)
	} else {
		details.WriteString("  (无)\n")
	}

	// Body预览
	details.WriteString("\n───────────────────────────────────────────────────────────\n")
	details.WriteString("【响应体 Body 预览】\n")
	bodyPreview := data.Body
	if len(bodyPreview) > 2000 {
		bodyPreview = bodyPreview[:2000] + "\n... (已截断，共 " + fmt.Sprintf("%d", len(data.Body)) + " 字节)"
	}
	details.WriteString(bodyPreview)

	return &pb.ValidateFingerprintResp{
		Success: true,
		Message: "验证完成",
		Matched: matched,
		Details: details.String(),
	}, nil
}

// validateAllFingerprints 批量验证所有指纹
func (l *TaskLogic) validateAllFingerprints(url, scope string, data *FingerprintData, startTime time.Time) (*pb.ValidateFingerprintResp, error) {
	// 构建查询条件
	filter := bson.M{"enabled": true}
	switch scope {
	case "builtin":
		filter["is_builtin"] = true
	case "custom":
		filter["is_builtin"] = false
	}

	// 获取所有启用的指纹（不分页）
	fingerprints, err := l.svcCtx.FingerprintModel.Find(l.ctx, filter, 0, 0)
	if err != nil {
		return &pb.ValidateFingerprintResp{
			Success: false,
			Message: "获取指纹列表失败: " + err.Error(),
		}, nil
	}

	// 遍历匹配
	var matched []*pb.MatchedFingerprintInfo
	for _, fp := range fingerprints {
		fpCopy := fp // 创建副本避免指针问题
		engine := l.newSingleFingerprintEngine(&fpCopy)
		isMatched, conditions := engine.MatchWithDetails(data)
		if isMatched {
			condStr := ""
			if len(conditions) > 0 {
				condStr = strings.Join(conditions, "; ")
			}
			matched = append(matched, &pb.MatchedFingerprintInfo{
				Id:                fp.Id.Hex(),
				Name:              fp.Name,
				Category:          fp.Category,
				IsBuiltin:         fp.IsBuiltin,
				MatchedConditions: condStr,
			})
		}
	}

	duration := time.Since(startTime)
	durationStr := fmt.Sprintf("%.2fs", duration.Seconds())

	return &pb.ValidateFingerprintResp{
		Success:      true,
		Message:      "验证完成",
		MatchedCount: int32(len(matched)),
		Duration:     durationStr,
		MatchedList:  matched,
	}, nil
}


// ==================== 指纹匹配辅助函数 ====================

// FingerprintData 用于指纹匹配的数据
type FingerprintData struct {
	Title        string
	Body         string
	BodyBytes    []byte
	Headers      map[string][]string
	HeaderString string
	Server       string
	URL          string
	FaviconHash  string
	Cookies      string
}

// fetchFingerprintData 请求URL获取指纹匹配数据
func (l *TaskLogic) fetchFingerprintData(targetUrl string) (*FingerprintData, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(targetUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 尝试检测并转换编码
	body := l.decodeBody(bodyBytes)

	// 提取标题
	title := ""
	titleRe := regexp.MustCompile(`(?i)<title[^>]*>([^<]*)</title>`)
	if matches := titleRe.FindStringSubmatch(body); len(matches) > 1 {
		title = strings.TrimSpace(matches[1])
	}

	// 构建header字符串
	var headerStr strings.Builder
	for key, values := range resp.Header {
		for _, v := range values {
			headerStr.WriteString(key)
			headerStr.WriteString(": ")
			headerStr.WriteString(v)
			headerStr.WriteString("\n")
		}
	}

	// 获取favicon并计算MMH3 hash
	faviconHash := l.fetchFaviconHash(targetUrl, body, client)

	return &FingerprintData{
		Title:        title,
		Body:         body,
		BodyBytes:    bodyBytes,
		Headers:      resp.Header,
		HeaderString: headerStr.String(),
		Server:       resp.Header.Get("Server"),
		URL:          targetUrl,
		FaviconHash:  faviconHash,
		Cookies:      resp.Header.Get("Set-Cookie"),
	}, nil
}

// decodeBody 尝试检测并转换编码
func (l *TaskLogic) decodeBody(bodyBytes []byte) string {
	// 先尝试UTF-8
	body := string(bodyBytes)

	// 检测是否包含GBK/GB2312编码标识
	if strings.Contains(strings.ToLower(body), "charset=gb") ||
		strings.Contains(strings.ToLower(body), "charset=\"gb") {
		// 尝试GBK解码
		reader := transform.NewReader(bytes.NewReader(bodyBytes), simplifiedchinese.GBK.NewDecoder())
		decoded, err := io.ReadAll(reader)
		if err == nil {
			return string(decoded)
		}
	}

	return body
}

// fetchFaviconHash 获取favicon并计算MMH3 hash
func (l *TaskLogic) fetchFaviconHash(baseUrl, body string, client *http.Client) string {
	// 尝试从HTML中提取favicon路径
	faviconUrl := ""

	// 1. 尝试从link标签获取
	linkRe := regexp.MustCompile(`(?i)<link[^>]*rel=["'](?:shortcut )?icon["'][^>]*href=["']([^"']+)["']`)
	if matches := linkRe.FindStringSubmatch(body); len(matches) > 1 {
		faviconUrl = matches[1]
	}
	// 也尝试href在rel前面的情况
	if faviconUrl == "" {
		linkRe2 := regexp.MustCompile(`(?i)<link[^>]*href=["']([^"']+)["'][^>]*rel=["'](?:shortcut )?icon["']`)
		if matches := linkRe2.FindStringSubmatch(body); len(matches) > 1 {
			faviconUrl = matches[1]
		}
	}

	// 2. 如果没找到，使用默认路径
	if faviconUrl == "" {
		faviconUrl = "/favicon.ico"
	}

	// 3. 处理相对路径
	if !strings.HasPrefix(faviconUrl, "http") {
		if strings.HasPrefix(faviconUrl, "//") {
			faviconUrl = "https:" + faviconUrl
		} else if strings.HasPrefix(faviconUrl, "/") {
			u := l.parseBaseUrl(baseUrl)
			if u != "" {
				faviconUrl = u + faviconUrl
			}
		} else {
			u := l.parseBaseUrl(baseUrl)
			if u != "" {
				faviconUrl = u + "/" + faviconUrl
			}
		}
	}

	// 4. 请求favicon
	resp, err := client.Get(faviconUrl)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	faviconBytes, err := io.ReadAll(resp.Body)
	if err != nil || len(faviconBytes) == 0 {
		return ""
	}

	// 5. 计算MMH3 hash (Shodan风格)
	return l.calculateMMH3Hash(faviconBytes)
}

// parseBaseUrl 解析URL获取基础部分 (scheme://host:port)
func (l *TaskLogic) parseBaseUrl(rawUrl string) string {
	if idx := strings.Index(rawUrl, "://"); idx > 0 {
		scheme := rawUrl[:idx]
		rest := rawUrl[idx+3:]
		if slashIdx := strings.Index(rest, "/"); slashIdx > 0 {
			return scheme + "://" + rest[:slashIdx]
		}
		return scheme + "://" + rest
	}
	return ""
}

// calculateMMH3Hash 计算Shodan风格的MMH3 favicon hash
func (l *TaskLogic) calculateMMH3Hash(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	b64 := base64.StdEncoding.EncodeToString(data)

	// 添加换行符（每76字符）模拟标准base64输出
	var b64WithNewlines strings.Builder
	for i := 0; i < len(b64); i += 76 {
		end := i + 76
		if end > len(b64) {
			end = len(b64)
		}
		b64WithNewlines.WriteString(b64[i:end])
		b64WithNewlines.WriteString("\n")
	}

	hash := l.mmh3Hash32([]byte(b64WithNewlines.String()))
	return fmt.Sprintf("%d", int32(hash))
}

// mmh3Hash32 MurmurHash3 32位实现
func (l *TaskLogic) mmh3Hash32(data []byte) uint32 {
	const (
		c1 = 0xcc9e2d51
		c2 = 0x1b873593
		r1 = 15
		r2 = 13
		m  = 5
		n  = 0xe6546b64
	)

	length := len(data)
	h := uint32(0)

	nblocks := length / 4
	for i := 0; i < nblocks; i++ {
		k := uint32(data[i*4]) | uint32(data[i*4+1])<<8 | uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2

		h ^= k
		h = (h << r2) | (h >> (32 - r2))
		h = h*m + n
	}

	tail := data[nblocks*4:]
	var k uint32
	switch len(tail) {
	case 3:
		k ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k ^= uint32(tail[0])
		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2
		h ^= k
	}

	h ^= uint32(length)
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16

	return h
}

// truncateString 截断字符串
func (l *TaskLogic) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ==================== 指纹匹配引擎 ====================

// SingleFingerprintEngine 单指纹匹配引擎
type SingleFingerprintEngine struct {
	fp *model.Fingerprint
}

func (l *TaskLogic) newSingleFingerprintEngine(fp *model.Fingerprint) *SingleFingerprintEngine {
	return &SingleFingerprintEngine{fp: fp}
}

func (e *SingleFingerprintEngine) Match(data *FingerprintData) bool {
	matched, _ := e.MatchWithDetails(data)
	return matched
}

// MatchWithDetails 执行匹配并返回匹配的条件详情
func (e *SingleFingerprintEngine) MatchWithDetails(data *FingerprintData) (bool, []string) {
	fp := e.fp

	// 优先使用Rule字段（ARL格式规则语法）
	if fp.Rule != "" {
		return matchRuleWithDetails(fp.Rule, data)
	}

	// 使用Wappalyzer格式规则
	return matchWappalyzerRulesWithDetails(fp, data)
}

// matchRuleWithDetails 匹配自定义格式规则并返回匹配的条件
func matchRuleWithDetails(rule string, data *FingerprintData) (bool, []string) {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return false, nil
	}

	// 处理OR逻辑 (||)
	parts := splitByOperator(rule, "||")
	if len(parts) > 1 {
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			matched, conditions := matchRuleAndWithDetails(part, data)
			if matched {
				return true, conditions
			}
		}
		return false, nil
	}

	return matchRuleAndWithDetails(rule, data)
}

func matchRuleAndWithDetails(rule string, data *FingerprintData) (bool, []string) {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return false, nil
	}

	var matchedConditions []string

	parts := splitByOperator(rule, "&&")
	if len(parts) > 1 {
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			matched, detail := matchSingleConditionWithDetails(part, data)
			if !matched {
				return false, nil
			}
			matchedConditions = append(matchedConditions, detail)
		}
		return true, matchedConditions
	}

	matched, detail := matchSingleConditionWithDetails(rule, data)
	if matched {
		return true, []string{detail}
	}
	return false, nil
}

func splitByOperator(rule, op string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(rule); i++ {
		c := rule[i]
		if (c == '"' || c == '\'') && (i == 0 || rule[i-1] != '\\') {
			if !inQuote {
				inQuote = true
				quoteChar = c
			} else if c == quoteChar {
				inQuote = false
			}
		}
		if !inQuote && i+len(op) <= len(rule) && rule[i:i+len(op)] == op {
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
			i += len(op) - 1
			continue
		}
		current.WriteByte(c)
	}
	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}
	return parts
}

// matchSingleConditionWithDetails 匹配单个条件并返回详情
func matchSingleConditionWithDetails(condition string, data *FingerprintData) (bool, string) {
	condition = strings.TrimSpace(condition)

	var condType, value string
	var negate bool

	if idx := strings.Index(condition, "!=\""); idx > 0 {
		condType = strings.TrimSpace(condition[:idx])
		value = extractQuotedValue(condition[idx+3:])
		negate = true
	} else if idx := strings.Index(condition, "=\""); idx > 0 {
		condType = strings.TrimSpace(condition[:idx])
		value = extractQuotedValue(condition[idx+2:])
		negate = false
	} else if idx := strings.Index(condition, "="); idx > 0 {
		condType = strings.TrimSpace(condition[:idx])
		value = strings.Trim(strings.TrimSpace(condition[idx+1:]), "\"'")
		negate = false
	} else {
		return false, ""
	}

	var result bool
	var matchedValue string
	condTypeLower := strings.ToLower(condType)

	switch condTypeLower {
	case "body":
		result = containsIgnoreCase(data.Body, value)
		if result {
			matchedValue = findMatchContext(data.Body, value, 50)
		}
	case "title":
		result = containsIgnoreCase(data.Title, value)
		if result {
			matchedValue = data.Title
		}
	case "header":
		result = containsIgnoreCase(data.HeaderString, value)
		if result {
			matchedValue = findMatchContext(data.HeaderString, value, 100)
		}
	case "server":
		result = containsIgnoreCase(data.Server, value)
		if result {
			matchedValue = data.Server
		}
	case "url":
		result = containsIgnoreCase(data.URL, value)
		if result {
			matchedValue = data.URL
		}
	case "cookie":
		result = containsIgnoreCase(data.Cookies, value)
		if result {
			matchedValue = findMatchContext(data.Cookies, value, 100)
		}
	case "icon_hash", "favicon_hash":
		result = data.FaviconHash == value
		if result {
			matchedValue = data.FaviconHash
		}
	default:
		return false, ""
	}

	if negate {
		result = !result
	}

	var detail string
	if result {
		if negate {
			detail = fmt.Sprintf("%s != \"%s\"", condType, value)
		} else {
			detail = fmt.Sprintf("%s = \"%s\" → 匹配到: %s", condType, value, truncateStr(matchedValue, 80))
		}
	}

	return result, detail
}

func extractQuotedValue(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return s
	}
	if s[0] == '"' || s[0] == '\'' {
		quoteChar := s[0]
		for i := 1; i < len(s); i++ {
			if s[i] == quoteChar && (i <= 1 || s[i-1] != '\\') {
				return s[1:i]
			}
		}
		return s[1:]
	}
	if s[len(s)-1] == '"' || s[len(s)-1] == '\'' {
		return s[:len(s)-1]
	}
	return s
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func findMatchContext(text, keyword string, contextLen int) string {
	textLower := strings.ToLower(text)
	keywordLower := strings.ToLower(keyword)

	idx := strings.Index(textLower, keywordLower)
	if idx < 0 {
		return ""
	}

	start := idx - contextLen
	if start < 0 {
		start = 0
	}
	end := idx + len(keyword) + contextLen
	if end > len(text) {
		end = len(text)
	}

	result := text[start:end]
	result = strings.ReplaceAll(result, "\n", " ")
	result = strings.ReplaceAll(result, "\r", "")

	prefix := ""
	suffix := ""
	if start > 0 {
		prefix = "..."
	}
	if end < len(text) {
		suffix = "..."
	}

	return prefix + result + suffix
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ==================== Wappalyzer规则匹配 ====================

// matchWappalyzerRulesWithDetails 匹配Wappalyzer格式规则并返回匹配详情
func matchWappalyzerRulesWithDetails(fp *model.Fingerprint, data *FingerprintData) (bool, []string) {
	hasRule := false
	allMatch := true
	var matchedConditions []string

	// Headers匹配
	if len(fp.Headers) > 0 {
		hasRule = true
		headerMatch := false
		for key, pattern := range fp.Headers {
			for hKey, hVal := range data.Headers {
				if strings.EqualFold(hKey, key) {
					headerValue := strings.Join(hVal, " ")
					if pattern == "" {
						headerMatch = true
						matchedConditions = append(matchedConditions, fmt.Sprintf("header[%s] 存在 → 匹配到: %s", key, truncateStr(headerValue, 80)))
						break
					}
					if matchRegexOrContains(headerValue, pattern) {
						headerMatch = true
						matchedConditions = append(matchedConditions, fmt.Sprintf("header[%s] =~ \"%s\" → 匹配到: %s", key, truncateStr(pattern, 50), truncateStr(headerValue, 80)))
						break
					}
				}
			}
			if headerMatch {
				break
			}
		}
		if !headerMatch {
			allMatch = false
		}
	}

	// HTML匹配
	if len(fp.HTML) > 0 && allMatch {
		hasRule = true
		htmlMatch := false
		for _, pattern := range fp.HTML {
			if matchRegexOrContains(data.Body, pattern) {
				htmlMatch = true
				matchedConditions = append(matchedConditions, fmt.Sprintf("html =~ \"%s\" → 匹配到", truncateStr(pattern, 50)))
				break
			}
		}
		if !htmlMatch {
			allMatch = false
		}
	}

	// Scripts匹配
	if len(fp.Scripts) > 0 && allMatch {
		hasRule = true
		scriptMatch := false
		scriptSrcRe := regexp.MustCompile(`(?i)<script[^>]*src=["']([^"']+)["']`)
		scriptSrcs := scriptSrcRe.FindAllStringSubmatch(data.Body, -1)
		for _, pattern := range fp.Scripts {
			for _, src := range scriptSrcs {
				if len(src) > 1 && matchRegexOrContains(src[1], pattern) {
					scriptMatch = true
					matchedConditions = append(matchedConditions, fmt.Sprintf("scripts =~ \"%s\" → 匹配到: %s", truncateStr(pattern, 50), truncateStr(src[1], 80)))
					break
				}
			}
			if scriptMatch {
				break
			}
		}
		if !scriptMatch {
			allMatch = false
		}
	}

	// Cookies匹配
	if len(fp.Cookies) > 0 && allMatch {
		hasRule = true
		cookieMatch := false
		for key, pattern := range fp.Cookies {
			if containsIgnoreCase(data.Cookies, key) {
				if pattern == "" || matchRegexOrContains(data.Cookies, pattern) {
					cookieMatch = true
					matchedConditions = append(matchedConditions, fmt.Sprintf("cookie[%s] =~ \"%s\" → 匹配到", key, pattern))
					break
				}
			}
		}
		if !cookieMatch {
			allMatch = false
		}
	}

	// Meta匹配
	if len(fp.Meta) > 0 && allMatch {
		hasRule = true
		metaMatch := false
		for key, pattern := range fp.Meta {
			metaPatterns := []string{
				fmt.Sprintf(`(?i)<meta[^>]*name=["']?%s["']?[^>]*content=["']([^"']*)["']`, regexp.QuoteMeta(key)),
				fmt.Sprintf(`(?i)<meta[^>]*content=["']([^"']*)["'][^>]*name=["']?%s["']?`, regexp.QuoteMeta(key)),
			}
			for _, mp := range metaPatterns {
				re := regexp.MustCompile(mp)
				if matches := re.FindStringSubmatch(data.Body); len(matches) > 1 {
					if pattern == "" || matchRegexOrContains(matches[1], pattern) {
						metaMatch = true
						matchedConditions = append(matchedConditions, fmt.Sprintf("meta[%s] =~ \"%s\" → 匹配到: %s", key, pattern, truncateStr(matches[1], 80)))
						break
					}
				}
			}
			if metaMatch {
				break
			}
		}
		if !metaMatch {
			allMatch = false
		}
	}

	// CSS匹配
	if len(fp.CSS) > 0 && allMatch {
		hasRule = true
		cssMatch := false
		for _, pattern := range fp.CSS {
			if matchRegexOrContains(data.Body, pattern) {
				cssMatch = true
				matchedConditions = append(matchedConditions, fmt.Sprintf("css =~ \"%s\" → 匹配到", truncateStr(pattern, 50)))
				break
			}
		}
		if !cssMatch {
			allMatch = false
		}
	}

	// URL匹配
	if len(fp.URL) > 0 && allMatch {
		hasRule = true
		urlMatch := false
		for _, pattern := range fp.URL {
			if matchRegexOrContains(data.URL, pattern) {
				urlMatch = true
				matchedConditions = append(matchedConditions, fmt.Sprintf("url =~ \"%s\" → 匹配到: %s", truncateStr(pattern, 50), data.URL))
				break
			}
		}
		if !urlMatch {
			allMatch = false
		}
	}

	if hasRule && allMatch {
		return true, matchedConditions
	}
	return false, nil
}

// matchRegexOrContains 尝试正则匹配，如果正则无效则回退到字符串包含匹配
func matchRegexOrContains(text, pattern string) bool {
	if pattern == "" {
		return true
	}
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return containsIgnoreCase(text, pattern)
	}
	return re.MatchString(text)
}
// ==================== POC验证 ====================

// ValidatePoc POC验证 - 通过任务队列下发给worker执行
func (l *TaskLogic) ValidatePoc(in *pb.ValidatePocReq) (*pb.ValidatePocResp, error) {
	l.Logger.Infof("ValidatePoc: url=%s, pocId=%s, pocType=%s", in.Url, in.PocId, in.PocType)

	if in.Url == "" {
		return &pb.ValidatePocResp{
			Success: false,
			Message: "URL不能为空",
		}, nil
	}

	// 生成任务ID
	taskId := fmt.Sprintf("poc-validate-%d", time.Now().UnixNano())

	// 构建POC验证任务配置 - 只传递pocId，Worker通过RPC获取POC内容
	taskConfig := map[string]interface{}{
		"taskType":    "poc_validate",
		"url":         in.Url,
		"pocId":       in.PocId,
		"pocType":     in.PocType,
		"timeout":     in.Timeout,
		"useTemplate": in.UseTemplate,
		"useCustom":   in.UseCustom,
	}

	if len(in.Severities) > 0 {
		taskConfig["severities"] = in.Severities
	}
	if len(in.Tags) > 0 {
		taskConfig["tags"] = in.Tags
	}

	configBytes, _ := json.Marshal(taskConfig)

	// 创建任务信息
	taskInfo := map[string]interface{}{
		"taskId":      taskId,
		"mainTaskId":  taskId, // 对于POC验证，使用相同的ID
		"workspaceId": "default", // 默认工作空间
		"taskName":    fmt.Sprintf("POC验证: %s", in.Url),
		"config":      string(configBytes),
		"priority":    10, // 高优先级
	}

	// 将任务推送到Redis队列
	taskInfoJson, _ := json.Marshal(taskInfo)
	queueKey := "cscan:task:queue"
	// 使用ZAdd添加到有序集合，分数为优先级（数字越小优先级越高）
	err := l.svcCtx.RedisClient.ZAdd(l.ctx, queueKey, redis.Z{
		Score:  10, // 高优先级
		Member: string(taskInfoJson),
	}).Err()

	if err != nil {
		l.Logger.Errorf("Failed to push POC validation task to queue: %v", err)
		return &pb.ValidatePocResp{
			Success: false,
			Message: "任务入队失败: " + err.Error(),
		}, nil
	}

	// 保存任务信息到Redis，用于后续查询结果
	taskInfoKey := fmt.Sprintf("cscan:task:info:%s", taskId)
	taskInfoData := map[string]string{
		"workspaceId": "default",
		"mainTaskId":  taskId,
		"status":      "PENDING",
		"createTime":  time.Now().Format("2006-01-02 15:04:05"),
	}
	taskInfoBytes, _ := json.Marshal(taskInfoData)
	l.svcCtx.RedisClient.Set(l.ctx, taskInfoKey, taskInfoBytes, 24*time.Hour)

	l.Logger.Infof("POC validation task queued: taskId=%s, url=%s", taskId, in.Url)

	return &pb.ValidatePocResp{
		Success: true,
		Message: "POC验证任务已下发，请稍后查询结果",
		Details: fmt.Sprintf("任务ID: %s\n目标URL: %s\nPOC类型: %s", taskId, in.Url, in.PocType),
		TaskId:  taskId,
	}, nil
}

// BatchValidatePoc 批量POC验证 - 通过任务队列下发给worker执行
func (l *TaskLogic) BatchValidatePoc(in *pb.BatchValidatePocReq) (*pb.BatchValidatePocResp, error) {
	l.Logger.Infof("BatchValidatePoc: urls=%d, pocType=%s", len(in.Urls), in.PocType)

	if len(in.Urls) == 0 {
		return &pb.BatchValidatePocResp{
			Success: false,
			Message: "URL列表不能为空",
		}, nil
	}

	startTime := time.Now()
	batchId := fmt.Sprintf("poc-batch-%d", startTime.UnixNano())
	var taskIds []string

	// 为每个URL创建一个验证任务
	for i, url := range in.Urls {
		taskId := fmt.Sprintf("%s-%d", batchId, i)
		taskIds = append(taskIds, taskId)

		// 构建任务配置
		taskConfig := map[string]interface{}{
			"taskType":    "poc_validate",
			"url":         url,
			"pocType":     in.PocType,
			"timeout":     in.Timeout,
			"useTemplate": in.UseTemplate,
			"useCustom":   in.UseCustom,
			"batchId":     batchId,
			"batchIndex":  i,
			"batchTotal":  len(in.Urls),
		}

		if len(in.Severities) > 0 {
			taskConfig["severities"] = in.Severities
		}
		if len(in.Tags) > 0 {
			taskConfig["tags"] = in.Tags
		}

		configBytes, _ := json.Marshal(taskConfig)

		// 创建任务信息
		taskInfo := map[string]interface{}{
			"taskId":      taskId,
			"mainTaskId":  batchId,
			"workspaceId": "default",
			"taskName":    fmt.Sprintf("批量POC验证: %s", url),
			"config":      string(configBytes),
			"priority":    15, // 批量任务优先级稍低
		}

		// 推送到队列
		taskInfoJson, _ := json.Marshal(taskInfo)
		queueKey := "cscan:task:queue"
		
		err := l.svcCtx.RedisClient.ZAdd(l.ctx, queueKey, redis.Z{
			Score:  15,
			Member: string(taskInfoJson),
		}).Err()

		if err != nil {
			l.Logger.Errorf("Failed to push batch POC task %s to queue: %v", taskId, err)
			continue
		}

		// 保存任务信息
		taskInfoKey := fmt.Sprintf("cscan:task:info:%s", taskId)
		taskInfoData := map[string]string{
			"workspaceId": "default",
			"mainTaskId":  batchId,
			"status":      "PENDING",
			"createTime":  time.Now().Format("2006-01-02 15:04:05"),
			"batchId":     batchId,
		}
		taskInfoBytes, _ := json.Marshal(taskInfoData)
		l.svcCtx.RedisClient.Set(l.ctx, taskInfoKey, taskInfoBytes, 24*time.Hour)
	}

	// 保存批次信息
	batchInfoKey := fmt.Sprintf("cscan:batch:info:%s", batchId)
	batchInfo := map[string]interface{}{
		"batchId":    batchId,
		"taskIds":    taskIds,
		"totalUrls":  len(in.Urls),
		"pocType":    in.PocType,
		"status":     "PENDING",
		"createTime": time.Now().Format("2006-01-02 15:04:05"),
	}
	batchInfoBytes, _ := json.Marshal(batchInfo)
	l.svcCtx.RedisClient.Set(l.ctx, batchInfoKey, batchInfoBytes, 24*time.Hour)

	duration := time.Since(startTime)
	l.Logger.Infof("Batch POC validation tasks queued: batchId=%s, tasks=%d, duration=%v", 
		batchId, len(taskIds), duration)

	return &pb.BatchValidatePocResp{
		Success:   true,
		Message:   "批量POC验证任务已下发",
		TotalUrls: int32(len(in.Urls)),
		Duration:  fmt.Sprintf("%.2fs", duration.Seconds()),
		BatchId:   batchId,
	}, nil
}

// validateSinglePoc 验证单个POC
func (l *TaskLogic) validateSinglePoc(pocId, url, pocType string, timeout int, startTime time.Time) (*pb.ValidatePocResp, error) {
	var poc *PocInfo

	if pocType == "custom" {
		// 获取自定义POC
		customPoc, err := l.svcCtx.CustomPocModel.FindById(l.ctx, pocId)
		if err != nil {
			return &pb.ValidatePocResp{
				Success: false,
				Message: "POC不存在",
			}, nil
		}
		poc = &PocInfo{
			Id:         customPoc.Id.Hex(),
			Name:       customPoc.Name,
			TemplateId: customPoc.TemplateId,
			Severity:   customPoc.Severity,
			Tags:       customPoc.Tags,
			Content:    customPoc.Content,
			PocType:    "custom",
		}
	} else {
		// 获取Nuclei模板
		template, err := l.svcCtx.NucleiTemplateModel.FindByTemplateId(l.ctx, pocId)
		if err != nil {
			return &pb.ValidatePocResp{
				Success: false,
				Message: "模板不存在",
			}, nil
		}
		poc = &PocInfo{
			Id:         template.TemplateId,
			Name:       template.Name,
			TemplateId: template.TemplateId,
			Severity:   template.Severity,
			Tags:       template.Tags,
			Content:    template.Content,
			PocType:    "nuclei",
		}
	}

	// 执行POC验证
	result := l.executePocValidation(url, poc, timeout)

	duration := time.Since(startTime)
	durationStr := fmt.Sprintf("%.2fs", duration.Seconds())

	var details string
	if result != nil {
		details = result.Details
		if result.Output != "" {
			details += "\n\n输出信息:\n" + result.Output
		}
	}

	return &pb.ValidatePocResp{
		Success:  true,
		Message:  "验证完成",
		Matched:  result != nil && result.Matched,
		Duration: durationStr,
		Details:  details,
		Results:  []*pb.PocValidationResult{result},
	}, nil
}

// validateAllPocs 验证所有POC
func (l *TaskLogic) validateAllPocs(url, pocType string, severities, tags []string, timeout int, useTemplate, useCustom bool, startTime time.Time) (*pb.ValidatePocResp, error) {
	// 获取POC列表
	pocs, err := l.getPocList(pocType, severities, tags, useTemplate, useCustom)
	if err != nil {
		return &pb.ValidatePocResp{
			Success: false,
			Message: "获取POC列表失败: " + err.Error(),
		}, nil
	}

	if len(pocs) == 0 {
		return &pb.ValidatePocResp{
			Success: false,
			Message: "没有找到符合条件的POC",
		}, nil
	}

	// 执行验证
	var results []*pb.PocValidationResult
	matchedCount := 0

	for _, poc := range pocs {
		result := l.executePocValidation(url, poc, timeout)
		if result != nil {
			results = append(results, result)
			if result.Matched {
				matchedCount++
			}
		}
	}

	duration := time.Since(startTime)
	durationStr := fmt.Sprintf("%.2fs", duration.Seconds())

	return &pb.ValidatePocResp{
		Success:      true,
		Message:      "验证完成",
		MatchedCount: int32(matchedCount),
		TotalCount:   int32(len(pocs)),
		Duration:     durationStr,
		Results:      results,
	}, nil
}

// PocInfo POC信息结构
type PocInfo struct {
	Id         string
	Name       string
	TemplateId string
	Severity   string
	Tags       []string
	Content    string
	PocType    string
}

// getPocList 获取POC列表
func (l *TaskLogic) getPocList(pocType string, severities, tags []string, useTemplate, useCustom bool) ([]*PocInfo, error) {
	var pocs []*PocInfo

	// 获取Nuclei模板
	if (pocType == "all" || pocType == "nuclei") && useTemplate {
		filter := bson.M{"enabled": true}
		if len(severities) > 0 {
			filter["severity"] = bson.M{"$in": severities}
		}
		if len(tags) > 0 {
			filter["tags"] = bson.M{"$in": tags}
		}

		templates, err := l.svcCtx.NucleiTemplateModel.FindEnabledByFilter(l.ctx, filter)
		if err != nil {
			return nil, err
		}

		for _, t := range templates {
			pocs = append(pocs, &PocInfo{
				Id:         t.TemplateId,
				Name:       t.Name,
				TemplateId: t.TemplateId,
				Severity:   t.Severity,
				Tags:       t.Tags,
				Content:    t.Content,
				PocType:    "nuclei",
			})
		}
	}

	// 获取自定义POC
	if (pocType == "all" || pocType == "custom") && useCustom {
		customPocs, err := l.svcCtx.CustomPocModel.FindEnabled(l.ctx)
		if err != nil {
			return nil, err
		}

		for _, p := range customPocs {
			// 过滤严重级别
			if len(severities) > 0 {
				found := false
				for _, s := range severities {
					if strings.EqualFold(p.Severity, s) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// 过滤标签
			if len(tags) > 0 {
				found := false
				for _, tag := range tags {
					for _, pTag := range p.Tags {
						if strings.EqualFold(pTag, tag) {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					continue
				}
			}

			pocs = append(pocs, &PocInfo{
				Id:         p.Id.Hex(),
				Name:       p.Name,
				TemplateId: p.TemplateId,
				Severity:   p.Severity,
				Tags:       p.Tags,
				Content:    p.Content,
				PocType:    "custom",
			})
		}
	}

	return pocs, nil
}

// executePocValidation 执行POC验证
func (l *TaskLogic) executePocValidation(url string, poc *PocInfo, timeout int) *pb.PocValidationResult {
	l.Logger.Infof("Executing POC validation: url=%s, poc=%s", url, poc.Name)

	// 这里应该调用nuclei执行POC验证
	// 由于nuclei需要外部命令执行，这里提供一个模拟实现
	// 实际生产环境中需要调用nuclei命令行工具

	result := &pb.PocValidationResult{
		PocId:      poc.Id,
		PocName:    poc.Name,
		TemplateId: poc.TemplateId,
		Severity:   poc.Severity,
		MatchedUrl: url,
		PocType:    poc.PocType,
		Tags:       poc.Tags,
	}

	// 模拟POC验证逻辑
	// 实际实现应该：
	// 1. 将POC内容写入临时文件
	// 2. 调用nuclei命令执行验证
	// 3. 解析nuclei输出结果
	// 4. 返回验证结果

	// 简单的模拟：检查URL是否可访问
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		result.Matched = false
		result.Details = fmt.Sprintf("无法访问目标URL: %v", err)
		result.Output = ""
		return result
	}
	defer resp.Body.Close()

	// 模拟匹配逻辑（实际应该由nuclei执行）
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 简单的关键词匹配模拟
	matched := false
	var matchDetails []string

	// 检查一些常见的漏洞特征
	if strings.Contains(strings.ToLower(bodyStr), "error") ||
		strings.Contains(strings.ToLower(bodyStr), "exception") ||
		resp.StatusCode >= 500 {
		matched = true
		matchDetails = append(matchDetails, "检测到错误信息或异常状态码")
	}

	// 检查特定的POC关键词
	if strings.Contains(strings.ToLower(poc.Content), "sql") && 
		(strings.Contains(strings.ToLower(bodyStr), "mysql") || 
		 strings.Contains(strings.ToLower(bodyStr), "oracle") ||
		 strings.Contains(strings.ToLower(bodyStr), "postgresql")) {
		matched = true
		matchDetails = append(matchDetails, "可能存在SQL注入漏洞")
	}

	result.Matched = matched
	if matched {
		result.Details = fmt.Sprintf("POC验证成功\n目标: %s\n状态码: %d\n匹配条件: %s", 
			url, resp.StatusCode, strings.Join(matchDetails, ", "))
		result.Output = fmt.Sprintf("HTTP %d %s\nContent-Length: %d", 
			resp.StatusCode, resp.Status, len(body))
	} else {
		result.Details = fmt.Sprintf("POC验证失败\n目标: %s\n状态码: %d\n未发现漏洞特征", 
			url, resp.StatusCode)
		result.Output = fmt.Sprintf("HTTP %d %s", resp.StatusCode, resp.Status)
	}

	return result
}

// executeBatchValidation 执行批量验证
func (l *TaskLogic) executeBatchValidation(urls []string, pocs []*PocInfo, timeout, concurrency int) ([]*pb.PocValidationResult, map[string]int32) {
	var results []*pb.PocValidationResult
	urlStats := make(map[string]int32)

	// 初始化URL统计
	for _, url := range urls {
		urlStats[url] = 0
	}

	// 使用并发控制
	semaphore := make(chan struct{}, concurrency)
	resultChan := make(chan *pb.PocValidationResult, len(urls)*len(pocs))

	var totalTasks int
	for _, url := range urls {
		for _, poc := range pocs {
			totalTasks++
			go func(u string, p *PocInfo) {
				semaphore <- struct{}{} // 获取信号量
				defer func() { <-semaphore }() // 释放信号量

				result := l.executePocValidation(u, p, timeout)
				resultChan <- result
			}(url, poc)
		}
	}

	// 收集结果
	for i := 0; i < totalTasks; i++ {
		result := <-resultChan
		results = append(results, result)
		if result.Matched {
			urlStats[result.MatchedUrl]++
		}
	}

	return results, urlStats
}
// GetPocValidationResult 查询POC验证结果
func (l *TaskLogic) GetPocValidationResult(in *pb.GetPocValidationResultReq) (*pb.GetPocValidationResultResp, error) {
	l.Logger.Infof("GetPocValidationResult: taskId=%s, batchId=%s", in.TaskId, in.BatchId)

	if in.TaskId == "" && in.BatchId == "" {
		return &pb.GetPocValidationResultResp{
			Success: false,
			Message: "任务ID或批次ID不能为空",
		}, nil
	}

	if in.BatchId != "" {
		// 查询批次结果
		return l.getBatchValidationResult(in.BatchId)
	} else {
		// 查询单个任务结果
		return l.getSingleValidationResult(in.TaskId)
	}
}

// getSingleValidationResult 获取单个任务验证结果
func (l *TaskLogic) getSingleValidationResult(taskId string) (*pb.GetPocValidationResultResp, error) {
	// 从Redis获取任务信息
	taskInfoKey := fmt.Sprintf("cscan:task:info:%s", taskId)
	taskInfoData, err := l.svcCtx.RedisClient.Get(l.ctx, taskInfoKey).Result()
	if err != nil {
		return &pb.GetPocValidationResultResp{
			Success: false,
			Message: "任务不存在或已过期",
		}, nil
	}

	var taskInfo map[string]string
	if err := json.Unmarshal([]byte(taskInfoData), &taskInfo); err != nil {
		return &pb.GetPocValidationResultResp{
			Success: false,
			Message: "任务信息解析失败",
		}, nil
	}

	// 查询任务结果
	resultKey := fmt.Sprintf("cscan:task:result:%s", taskId)
	resultData, err := l.svcCtx.RedisClient.Get(l.ctx, resultKey).Result()
	
	var results []*pb.PocValidationResult
	status := taskInfo["status"]
	if status == "" {
		status = "PENDING"
	}

	if err == nil && resultData != "" {
		// 解析结果 - worker保存的格式是 {taskId, batchId, status, results, updateTime}
		var resultWrapper struct {
			TaskId     string `json:"taskId"`
			BatchId    string `json:"batchId"`
			Status     string `json:"status"`
			Error      string `json:"error"`
			UpdateTime string `json:"updateTime"`
			Results    []struct {
				PocId      string   `json:"pocId"`
				PocName    string   `json:"pocName"`
				TemplateId string   `json:"templateId"`
				Severity   string   `json:"severity"`
				Matched    bool     `json:"matched"`
				MatchedUrl string   `json:"matchedUrl"`
				Details    string   `json:"details"`
				Output     string   `json:"output"`
				PocType    string   `json:"pocType"`
				Tags       []string `json:"tags"`
			} `json:"results"`
		}
		
		if json.Unmarshal([]byte(resultData), &resultWrapper) == nil {
			// 更新状态
			if resultWrapper.Status != "" {
				status = resultWrapper.Status
			}
			
			// 转换结果
			for _, r := range resultWrapper.Results {
				results = append(results, &pb.PocValidationResult{
					PocId:      r.PocId,
					PocName:    r.PocName,
					TemplateId: r.TemplateId,
					Severity:   r.Severity,
					Matched:    r.Matched,
					MatchedUrl: r.MatchedUrl,
					Details:    r.Details,
					Output:     r.Output,
					PocType:    r.PocType,
					Tags:       r.Tags,
				})
			}
			
			// 更新taskInfo中的updateTime
			if resultWrapper.UpdateTime != "" {
				taskInfo["updateTime"] = resultWrapper.UpdateTime
			}
		}
	}

	return &pb.GetPocValidationResultResp{
		Success:        true,
		Message:        "查询成功",
		Status:         status,
		CompletedCount: int32(len(results)),
		TotalCount:     1,
		Results:        results,
		CreateTime:     taskInfo["createTime"],
		UpdateTime:     taskInfo["updateTime"],
	}, nil
}

// getBatchValidationResult 获取批次验证结果
func (l *TaskLogic) getBatchValidationResult(batchId string) (*pb.GetPocValidationResultResp, error) {
	// 从Redis获取批次信息
	batchInfoKey := fmt.Sprintf("cscan:batch:info:%s", batchId)
	batchInfoData, err := l.svcCtx.RedisClient.Get(l.ctx, batchInfoKey).Result()
	if err != nil {
		return &pb.GetPocValidationResultResp{
			Success: false,
			Message: "批次不存在或已过期",
		}, nil
	}

	var batchInfo map[string]interface{}
	if err := json.Unmarshal([]byte(batchInfoData), &batchInfo); err != nil {
		return &pb.GetPocValidationResultResp{
			Success: false,
			Message: "批次信息解析失败",
		}, nil
	}

	// 获取任务ID列表
	taskIdsInterface, ok := batchInfo["taskIds"].([]interface{})
	if !ok {
		return &pb.GetPocValidationResultResp{
			Success: false,
			Message: "批次任务列表格式错误",
		}, nil
	}

	var taskIds []string
	for _, id := range taskIdsInterface {
		if idStr, ok := id.(string); ok {
			taskIds = append(taskIds, idStr)
		}
	}

	// 查询所有任务的结果
	var results []*pb.PocValidationResult
	completedCount := 0

	for _, taskId := range taskIds {
		resultKey := fmt.Sprintf("cscan:task:result:%s", taskId)
		resultData, err := l.svcCtx.RedisClient.Get(l.ctx, resultKey).Result()
		
		if err == nil && resultData != "" {
			// 解析结果 - worker保存的格式
			var resultWrapper struct {
				Status  string `json:"status"`
				Results []struct {
					PocId      string   `json:"pocId"`
					PocName    string   `json:"pocName"`
					TemplateId string   `json:"templateId"`
					Severity   string   `json:"severity"`
					Matched    bool     `json:"matched"`
					MatchedUrl string   `json:"matchedUrl"`
					Details    string   `json:"details"`
					Output     string   `json:"output"`
					PocType    string   `json:"pocType"`
					Tags       []string `json:"tags"`
				} `json:"results"`
			}
			
			if json.Unmarshal([]byte(resultData), &resultWrapper) == nil {
				completedCount++
				for _, r := range resultWrapper.Results {
					results = append(results, &pb.PocValidationResult{
						PocId:      r.PocId,
						PocName:    r.PocName,
						TemplateId: r.TemplateId,
						Severity:   r.Severity,
						Matched:    r.Matched,
						MatchedUrl: r.MatchedUrl,
						Details:    r.Details,
						Output:     r.Output,
						PocType:    r.PocType,
						Tags:       r.Tags,
					})
				}
			}
		}
	}

	// 确定整体状态
	status := "PENDING"
	if completedCount == len(taskIds) {
		status = "SUCCESS"
	} else if completedCount > 0 {
		status = "STARTED"
	}

	createTime := ""
	if ct, ok := batchInfo["createTime"].(string); ok {
		createTime = ct
	}

	return &pb.GetPocValidationResultResp{
		Success:        true,
		Message:        "查询成功",
		Status:         status,
		CompletedCount: int32(completedCount),
		TotalCount:     int32(len(taskIds)),
		Results:        results,
		CreateTime:     createTime,
		UpdateTime:     time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}


// GetPocById 根据ID获取POC内容
func (l *TaskLogic) GetPocById(in *pb.GetPocByIdReq) (*pb.GetPocByIdResp, error) {
	l.Logger.Infof("GetPocById: pocId=%s, pocType=%s", in.PocId, in.PocType)

	if in.PocId == "" {
		return &pb.GetPocByIdResp{
			Success: false,
			Message: "POC ID不能为空",
		}, nil
	}

	if in.PocType == "custom" || in.PocType == "" {
		// 获取自定义POC
		poc, err := l.svcCtx.CustomPocModel.FindById(l.ctx, in.PocId)
		if err == nil && poc != nil {
			l.Logger.Infof("Found custom POC: %s, content length: %d", poc.Name, len(poc.Content))
			return &pb.GetPocByIdResp{
				Success:    true,
				Message:    "success",
				PocId:      poc.Id.Hex(),
				Name:       poc.Name,
				TemplateId: poc.TemplateId,
				Severity:   poc.Severity,
				Tags:       poc.Tags,
				Content:    poc.Content,
				PocType:    "custom",
			}, nil
		}
		l.Logger.Infof("Custom POC not found by id: %s, trying nuclei template", in.PocId)
	}

	if in.PocType == "nuclei" || in.PocType == "" {
		// 获取Nuclei模板
		template, err := l.svcCtx.NucleiTemplateModel.FindByTemplateId(l.ctx, in.PocId)
		if err == nil && template != nil {
			l.Logger.Infof("Found nuclei template: %s, content length: %d", template.Name, len(template.Content))
			return &pb.GetPocByIdResp{
				Success:    true,
				Message:    "success",
				PocId:      template.TemplateId,
				Name:       template.Name,
				TemplateId: template.TemplateId,
				Severity:   template.Severity,
				Tags:       template.Tags,
				Content:    template.Content,
				PocType:    "nuclei",
			}, nil
		}
		l.Logger.Infof("Nuclei template not found by id: %s", in.PocId)
	}

	return &pb.GetPocByIdResp{
		Success: false,
		Message: "POC不存在",
	}, nil
}

// GetTemplatesByIds 根据ID列表批量获取模板内容
func (l *TaskLogic) GetTemplatesByIds(in *pb.GetTemplatesByIdsReq) (*pb.GetTemplatesByIdsResp, error) {
	l.Logger.Infof("GetTemplatesByIds: nucleiTemplateIds=%d, customPocIds=%d", len(in.NucleiTemplateIds), len(in.CustomPocIds))

	var templates []string

	// 获取Nuclei模板内容
	if len(in.NucleiTemplateIds) > 0 {
		nucleiTemplates, err := l.svcCtx.NucleiTemplateModel.FindByIds(l.ctx, in.NucleiTemplateIds)
		if err != nil {
			l.Logger.Errorf("Failed to get nuclei templates by ids: %v", err)
		} else {
			for _, t := range nucleiTemplates {
				if t.Content != "" {
					templates = append(templates, t.Content)
				}
			}
			l.Logger.Infof("Fetched %d nuclei templates", len(nucleiTemplates))
		}
	}

	// 获取自定义POC内容
	if len(in.CustomPocIds) > 0 {
		customPocs, err := l.svcCtx.CustomPocModel.FindByIds(l.ctx, in.CustomPocIds)
		if err != nil {
			l.Logger.Errorf("Failed to get custom pocs by ids: %v", err)
		} else {
			for _, poc := range customPocs {
				if poc.Content != "" {
					templates = append(templates, poc.Content)
				}
			}
			l.Logger.Infof("Fetched %d custom POCs", len(customPocs))
		}
	}

	l.Logger.Infof("GetTemplatesByIds: total %d templates fetched", len(templates))

	return &pb.GetTemplatesByIdsResp{
		Success:   true,
		Message:   "success",
		Templates: templates,
		Count:     int32(len(templates)),
	}, nil
}

// GetHttpServiceMappings 获取HTTP服务映射
func (l *TaskLogic) GetHttpServiceMappings(in *pb.GetHttpServiceMappingsReq) (*pb.GetHttpServiceMappingsResp, error) {
	l.Logger.Infof("GetHttpServiceMappings: enabledOnly=%v", in.EnabledOnly)

	var docs []model.HttpServiceMapping
	var err error

	if in.EnabledOnly {
		docs, err = l.svcCtx.HttpServiceMappingModel.FindEnabled(l.ctx)
	} else {
		docs, err = l.svcCtx.HttpServiceMappingModel.FindAll(l.ctx)
	}

	if err != nil {
		l.Logger.Errorf("Failed to get http service mappings: %v", err)
		return &pb.GetHttpServiceMappingsResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	var pbMappings []*pb.HttpServiceMappingDocument
	for _, doc := range docs {
		pbMappings = append(pbMappings, &pb.HttpServiceMappingDocument{
			Id:          doc.Id.Hex(),
			ServiceName: doc.ServiceName,
			IsHttp:      doc.IsHttp,
			Description: doc.Description,
			Enabled:     doc.Enabled,
		})
	}

	l.Logger.Infof("Found %d http service mappings", len(pbMappings))

	return &pb.GetHttpServiceMappingsResp{
		Success:  true,
		Message:  "success",
		Mappings: pbMappings,
		Count:    int32(len(pbMappings)),
	}, nil
}
