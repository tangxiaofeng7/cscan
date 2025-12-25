package logic

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cscan/api/internal/svc"
	"cscan/api/internal/types"

	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

// ReportDetailLogic 报告详情
type ReportDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReportDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportDetailLogic {
	return &ReportDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReportDetailLogic) ReportDetail(req *types.ReportDetailReq, workspaceId string) (*types.ReportDetailResp, error) {
	l.Logger.Infof("ReportDetail: taskId=%s, workspaceId=%s", req.TaskId, workspaceId)
	
	// 获取任务信息 - 先在指定工作空间查找，找不到则在默认工作空间查找
	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)
	task, err := taskModel.FindById(l.ctx, req.TaskId)
	
	// 如果在当前工作空间找不到，尝试在默认工作空间查找
	actualWorkspaceId := workspaceId
	if err != nil && workspaceId != "" {
		l.Logger.Infof("Task not found in workspace %s, trying default workspace", workspaceId)
		taskModel = l.svcCtx.GetMainTaskModel("")
		task, err = taskModel.FindById(l.ctx, req.TaskId)
		if err == nil {
			actualWorkspaceId = ""
		}
	}
	
	// 如果还是找不到，尝试在默认工作空间查找（workspaceId 本身就是空的情况）
	if err != nil && workspaceId == "" {
		l.Logger.Infof("Task not found in default workspace, error: %v", err)
	}
	
	if err != nil {
		l.Logger.Errorf("FindById failed: %v", err)
		return &types.ReportDetailResp{Code: 400, Msg: "任务不存在"}, nil
	}
	
	// 确定用于查询资产的 taskId
	// 资产保存时使用的是 task.Id.Hex() (MongoDB ObjectID) 作为 taskId
	queryTaskId := task.Id.Hex()
	l.Logger.Infof("Found task: name=%s, taskId(UUID)=%s, queryTaskId(ObjectID)=%s, actualWorkspaceId=%s", task.Name, task.TaskId, queryTaskId, actualWorkspaceId)

	// 获取资产列表 - 使用实际的工作空间
	assetModel := l.svcCtx.GetAssetModel(actualWorkspaceId)
	
	// 先尝试用 ObjectID 查询
	assets, err := assetModel.Find(l.ctx, bson.M{"taskId": queryTaskId}, 0, 0)
	if err != nil {
		l.Logger.Errorf("查询资产失败: %v", err)
	}
	l.Logger.Infof("Found %d assets for queryTaskId(ObjectID)=%s in workspace=%s", len(assets), queryTaskId, actualWorkspaceId)
	
	// 如果没找到，尝试用 UUID 查询（兼容旧数据）
	if len(assets) == 0 {
		l.Logger.Infof("No assets found with ObjectID, trying UUID: %s", task.TaskId)
		assets, err = assetModel.Find(l.ctx, bson.M{"taskId": task.TaskId}, 0, 0)
		if err != nil {
			l.Logger.Errorf("查询资产(UUID)失败: %v", err)
		}
		l.Logger.Infof("Found %d assets for taskId(UUID)=%s", len(assets), task.TaskId)
	}
	
	// 如果在当前工作空间找不到资产，尝试在默认工作空间查找
	if len(assets) == 0 && actualWorkspaceId != "" {
		l.Logger.Infof("No assets found in workspace %s, trying default workspace", actualWorkspaceId)
		defaultAssetModel := l.svcCtx.GetAssetModel("")
		assets, err = defaultAssetModel.Find(l.ctx, bson.M{"taskId": queryTaskId}, 0, 0)
		if err != nil {
			l.Logger.Errorf("查询默认工作空间资产失败: %v", err)
		}
		l.Logger.Infof("Found %d assets in default workspace with ObjectID", len(assets))
		
		// 也尝试用 UUID 查询默认工作空间
		if len(assets) == 0 {
			assets, err = defaultAssetModel.Find(l.ctx, bson.M{"taskId": task.TaskId}, 0, 0)
			if err != nil {
				l.Logger.Errorf("查询默认工作空间资产(UUID)失败: %v", err)
			}
			l.Logger.Infof("Found %d assets in default workspace with UUID", len(assets))
		}
	}

	// 获取漏洞列表 - 使用实际的工作空间
	vulModel := l.svcCtx.GetVulModel(actualWorkspaceId)
	vuls, err := vulModel.Find(l.ctx, bson.M{"task_id": queryTaskId}, 0, 0)
	if err != nil {
		l.Logger.Errorf("查询漏洞失败: %v", err)
	}
	l.Logger.Infof("Found %d vuls for queryTaskId(ObjectID)=%s in workspace=%s", len(vuls), queryTaskId, actualWorkspaceId)
	
	// 如果没找到，尝试用 UUID 查询（兼容旧数据）
	if len(vuls) == 0 {
		l.Logger.Infof("No vuls found with ObjectID, trying UUID: %s", task.TaskId)
		vuls, err = vulModel.Find(l.ctx, bson.M{"task_id": task.TaskId}, 0, 0)
		if err != nil {
			l.Logger.Errorf("查询漏洞(UUID)失败: %v", err)
		}
		l.Logger.Infof("Found %d vuls for taskId(UUID)=%s", len(vuls), task.TaskId)
	}
	
	// 如果在当前工作空间找不到漏洞，尝试在默认工作空间查找
	if len(vuls) == 0 && actualWorkspaceId != "" {
		l.Logger.Infof("No vuls found in workspace %s, trying default workspace", actualWorkspaceId)
		defaultVulModel := l.svcCtx.GetVulModel("")
		vuls, err = defaultVulModel.Find(l.ctx, bson.M{"task_id": queryTaskId}, 0, 0)
		if err != nil {
			l.Logger.Errorf("查询默认工作空间漏洞失败: %v", err)
		}
		l.Logger.Infof("Found %d vuls in default workspace with ObjectID", len(vuls))
		
		// 也尝试用 UUID 查询默认工作空间
		if len(vuls) == 0 {
			vuls, err = defaultVulModel.Find(l.ctx, bson.M{"task_id": task.TaskId}, 0, 0)
			if err != nil {
				l.Logger.Errorf("查询默认工作空间漏洞(UUID)失败: %v", err)
			}
			l.Logger.Infof("Found %d vuls in default workspace with UUID", len(vuls))
		}
	}

	// 统计信息
	portStats := make(map[int]int)
	serviceStats := make(map[string]int)
	appStats := make(map[string]int)
	severityStats := map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0, "info": 0}

	for _, asset := range assets {
		portStats[asset.Port]++
		if asset.Service != "" {
			serviceStats[asset.Service]++
		}
		for _, app := range asset.App {
			appStats[app]++
		}
	}

	for _, vul := range vuls {
		severity := strings.ToLower(vul.Severity)
		if _, ok := severityStats[severity]; ok {
			severityStats[severity]++
		}
	}

	// 转换资产列表
	assetList := make([]types.ReportAsset, 0, len(assets))
	for _, a := range assets {
		assetList = append(assetList, types.ReportAsset{
			Authority:  a.Authority,
			Host:       a.Host,
			Port:       a.Port,
			Service:    a.Service,
			Title:      a.Title,
			App:        a.App,
			HttpStatus: a.HttpStatus,
			Server:     a.Server,
			IconHash:   a.IconHash,
			Screenshot: a.Screenshot,
			CreateTime: a.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	// 转换漏洞列表
	vulList := make([]types.ReportVul, 0, len(vuls))
	for _, v := range vuls {
		vulList = append(vulList, types.ReportVul{
			Authority:  v.Authority,
			Url:        v.Url,
			PocFile:    v.PocFile,
			Severity:   v.Severity,
			Result:     v.Result,
			CreateTime: v.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	// 转换统计
	topPorts := make([]types.StatItem, 0)
	for port, count := range portStats {
		topPorts = append(topPorts, types.StatItem{Name: strconv.Itoa(port), Count: count})
	}
	topServices := make([]types.StatItem, 0)
	for svc, count := range serviceStats {
		topServices = append(topServices, types.StatItem{Name: svc, Count: count})
	}
	topApps := make([]types.StatItem, 0)
	for app, count := range appStats {
		topApps = append(topApps, types.StatItem{Name: app, Count: count})
	}

	return &types.ReportDetailResp{
		Code: 0,
		Msg:  "success",
		Data: &types.ReportData{
			TaskId:      req.TaskId,
			TaskName:    task.Name,
			Target:      task.Target,
			Status:      task.Status,
			CreateTime:  task.CreateTime.Format("2006-01-02 15:04:05"),
			AssetCount:  len(assets),
			VulCount:    len(vuls),
			Assets:      assetList,
			Vuls:        vulList,
			TopPorts:    topPorts,
			TopServices: topServices,
			TopApps:     topApps,
			VulStats:    severityStats,
		},
	}, nil
}

// ReportExportLogic 报告导出
type ReportExportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReportExportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportExportLogic {
	return &ReportExportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReportExportLogic) ReportExport(req *types.ReportExportReq, workspaceId string) ([]byte, string, error) {
	// 获取任务信息
	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)
	task, err := taskModel.FindById(l.ctx, req.TaskId)
	if err != nil {
		return nil, "", fmt.Errorf("任务不存在")
	}

	// 获取资产列表
	assetModel := l.svcCtx.GetAssetModel(workspaceId)
	assets, _ := assetModel.Find(l.ctx, bson.M{"taskId": task.TaskId}, 0, 0)

	// 获取漏洞列表
	vulModel := l.svcCtx.GetVulModel(workspaceId)
	vuls, _ := vulModel.Find(l.ctx, bson.M{"task_id": task.TaskId}, 0, 0)

	// 创建Excel文件
	f := excelize.NewFile()
	defer f.Close()

	// 概览Sheet
	f.SetSheetName("Sheet1", "概览")
	f.SetCellValue("概览", "A1", "扫描报告")
	f.SetCellValue("概览", "A3", "任务名称")
	f.SetCellValue("概览", "B3", task.Name)
	f.SetCellValue("概览", "A4", "扫描目标")
	f.SetCellValue("概览", "B4", task.Target)
	f.SetCellValue("概览", "A5", "任务状态")
	f.SetCellValue("概览", "B5", task.Status)
	f.SetCellValue("概览", "A6", "创建时间")
	f.SetCellValue("概览", "B6", task.CreateTime.Format("2006-01-02 15:04:05"))
	f.SetCellValue("概览", "A7", "资产数量")
	f.SetCellValue("概览", "B7", len(assets))
	f.SetCellValue("概览", "A8", "漏洞数量")
	f.SetCellValue("概览", "B8", len(vuls))

	// 设置概览样式
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle("概览", "A1", "A1", titleStyle)
	f.MergeCell("概览", "A1", "B1")

	// 资产Sheet
	f.NewSheet("资产列表")
	assetHeaders := []string{"地址", "主机", "端口", "服务", "标题", "应用", "状态码", "Server", "IconHash", "发现时间"}
	for i, h := range assetHeaders {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue("资产列表", cell, h)
	}
	for i, a := range assets {
		row := i + 2
		f.SetCellValue("资产列表", fmt.Sprintf("A%d", row), a.Authority)
		f.SetCellValue("资产列表", fmt.Sprintf("B%d", row), a.Host)
		f.SetCellValue("资产列表", fmt.Sprintf("C%d", row), a.Port)
		f.SetCellValue("资产列表", fmt.Sprintf("D%d", row), a.Service)
		f.SetCellValue("资产列表", fmt.Sprintf("E%d", row), a.Title)
		f.SetCellValue("资产列表", fmt.Sprintf("F%d", row), strings.Join(a.App, ", "))
		f.SetCellValue("资产列表", fmt.Sprintf("G%d", row), a.HttpStatus)
		f.SetCellValue("资产列表", fmt.Sprintf("H%d", row), a.Server)
		f.SetCellValue("资产列表", fmt.Sprintf("I%d", row), a.IconHash)
		f.SetCellValue("资产列表", fmt.Sprintf("J%d", row), a.CreateTime.Format("2006-01-02 15:04:05"))
	}

	// 漏洞Sheet
	f.NewSheet("漏洞列表")
	vulHeaders := []string{"地址", "URL", "POC", "严重级别", "结果", "发现时间"}
	for i, h := range vulHeaders {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue("漏洞列表", cell, h)
	}
	for i, v := range vuls {
		row := i + 2
		f.SetCellValue("漏洞列表", fmt.Sprintf("A%d", row), v.Authority)
		f.SetCellValue("漏洞列表", fmt.Sprintf("B%d", row), v.Url)
		f.SetCellValue("漏洞列表", fmt.Sprintf("C%d", row), v.PocFile)
		f.SetCellValue("漏洞列表", fmt.Sprintf("D%d", row), v.Severity)
		f.SetCellValue("漏洞列表", fmt.Sprintf("E%d", row), v.Result)
		f.SetCellValue("漏洞列表", fmt.Sprintf("F%d", row), v.CreateTime.Format("2006-01-02 15:04:05"))
	}

	// 设置列宽
	f.SetColWidth("资产列表", "A", "A", 30)
	f.SetColWidth("资产列表", "B", "B", 15)
	f.SetColWidth("资产列表", "E", "E", 40)
	f.SetColWidth("资产列表", "F", "F", 30)
	f.SetColWidth("漏洞列表", "A", "A", 30)
	f.SetColWidth("漏洞列表", "B", "B", 50)
	f.SetColWidth("漏洞列表", "C", "C", 40)
	f.SetColWidth("漏洞列表", "E", "E", 50)

	// 写入buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("report_%s_%s.xlsx", task.Name, time.Now().Format("20060102150405"))
	return buf.Bytes(), filename, nil
}
