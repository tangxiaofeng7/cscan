package logic

import (
	"context"
	"strconv"
	"strings"

	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/model"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

// parseQuerySyntax 解析查询语法
// 支持格式: port=80 && service=http || title="test"
func parseQuerySyntax(query string, filter bson.M) {
	query = strings.TrimSpace(query)
	if query == "" {
		return
	}

	// 简单解析：支持 field=value 格式，多个条件用 && 连接
	// 例如: port=80 && service=http && title=test
	conditions := strings.Split(query, "&&")
	for _, cond := range conditions {
		cond = strings.TrimSpace(cond)
		if cond == "" {
			continue
		}

		// 解析 field=value 或 field="value"
		parts := strings.SplitN(cond, "=", 2)
		if len(parts) != 2 {
			continue
		}

		field := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// 去除引号
		value = strings.Trim(value, "\"'")

		// 映射字段名
		switch strings.ToLower(field) {
		case "port":
			if port, err := strconv.Atoi(value); err == nil {
				filter["port"] = port
			}
		case "host", "ip":
			filter["host"] = bson.M{"$regex": value, "$options": "i"}
		case "service", "protocol":
			filter["service"] = bson.M{"$regex": value, "$options": "i"}
		case "title":
			filter["title"] = bson.M{"$regex": value, "$options": "i"}
		case "app", "finger", "fingerprint":
			filter["app"] = bson.M{"$regex": value, "$options": "i"}
		case "status", "httpstatus":
			filter["status"] = value
		case "domain":
			filter["domain"] = bson.M{"$regex": value, "$options": "i"}
		case "banner":
			filter["banner"] = bson.M{"$regex": value, "$options": "i"}
		}
	}
}

type AssetListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssetListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssetListLogic {
	return &AssetListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssetListLogic) AssetList(req *types.AssetListReq, workspaceId string) (resp *types.AssetListResp, err error) {
	assetModel := l.svcCtx.GetAssetModel(workspaceId)

	// 构建查询条件
	filter := bson.M{}

	// 如果有语法查询，解析语法
	if req.Query != "" {
		parseQuerySyntax(req.Query, filter)
	} else {
		// 快捷查询
		if req.Host != "" {
			filter["host"] = bson.M{"$regex": req.Host, "$options": "i"}
		}
		if req.Port > 0 {
			filter["port"] = req.Port
		}
		if req.Service != "" {
			filter["service"] = bson.M{"$regex": req.Service, "$options": "i"}
		}
		if req.Title != "" {
			filter["title"] = bson.M{"$regex": req.Title, "$options": "i"}
		}
		if req.App != "" {
			filter["app"] = bson.M{"$regex": req.App, "$options": "i"}
		}
		if req.HttpStatus != "" {
			filter["status"] = req.HttpStatus
		}
	}

	// 只看新资产
	if req.OnlyNew {
		filter["new"] = true
	}
	// 只看有更新
	if req.OnlyUpdated {
		filter["update"] = true
	}
	// 排除CDN/Cloud资产
	if req.ExcludeCdn {
		filter["cdn"] = bson.M{"$ne": true}
		filter["cloud"] = bson.M{"$ne": true}
	}

	// 查询总数
	total, err := assetModel.Count(l.ctx, filter)
	if err != nil {
		return &types.AssetListResp{Code: 500, Msg: "查询失败"}, nil
	}

	// 查询列表 - 支持按风险评分排序 
	var assets []model.Asset
	if req.SortByRisk {
		assets, err = assetModel.FindByRiskScore(l.ctx, filter, req.Page, req.PageSize, false)
	} else {
		sortField := "update_time"
		if !req.SortByUpdate {
			sortField = "create_time"
		}
		assets, err = assetModel.FindWithSort(l.ctx, filter, req.Page, req.PageSize, sortField)
	}
	if err != nil {
		return &types.AssetListResp{Code: 500, Msg: "查询失败"}, nil
	}

	// 转换响应
	list := make([]types.Asset, 0, len(assets))
	for _, a := range assets {
		// 获取归属地信息
		location := ""
		if len(a.Ip.IpV4) > 0 && a.Ip.IpV4[0].Location != "" {
			location = a.Ip.IpV4[0].Location
		}

		list = append(list, types.Asset{
			Id:         a.Id.Hex(),
			Authority:  a.Authority,
			Host:       a.Host,
			Port:       a.Port,
			Category:   a.Category,
			Service:    a.Service,
			Title:      a.Title,
			App:        a.App,
			HttpStatus: a.HttpStatus,
			HttpHeader: a.HttpHeader,
			HttpBody:   a.HttpBody,
			Banner:     a.Banner,
			IconHash:   a.IconHash,
			Screenshot: a.Screenshot,
			Location:   location,
			IsCDN:      a.IsCDN,
			IsCloud:    a.IsCloud,
			IsNew:      a.IsNewAsset,
			IsUpdated:  a.IsUpdated,
			CreateTime: a.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateTime: a.UpdateTime.Format("2006-01-02 15:04:05"),
			// 新增字段 - 风险评分 
			RiskScore: a.RiskScore,
			RiskLevel: a.RiskLevel,
		})
	}

	return &types.AssetListResp{
		Code:  0,
		Msg:   "success",
		Total: int(total),
		List:  list,
	}, nil
}

type AssetStatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssetStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssetStatLogic {
	return &AssetStatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssetStatLogic) AssetStat(workspaceId string) (resp *types.AssetStatResp, err error) {
	assetModel := l.svcCtx.GetAssetModel(workspaceId)

	// 总资产数
	totalAsset, _ := assetModel.Count(l.ctx, bson.M{})

	// 统计主机数（去重host）
	// 简化处理，实际应使用distinct
	totalHost := totalAsset

	// 新资产数
	newCount, _ := assetModel.Count(l.ctx, bson.M{"new": true})

	// 有更新的资产数
	updatedCount, _ := assetModel.Count(l.ctx, bson.M{"update": true})

	// Top端口
	portStats, _ := assetModel.AggregatePort(l.ctx, 10)
	topPorts := make([]types.StatItem, 0, len(portStats))
	for _, s := range portStats {
		topPorts = append(topPorts, types.StatItem{
			Name:  strconv.Itoa(s.Port),
			Count: s.Count,
		})
	}

	// Top服务
	serviceStats, _ := assetModel.Aggregate(l.ctx, "service", 10)
	topService := make([]types.StatItem, 0, len(serviceStats))
	for _, s := range serviceStats {
		topService = append(topService, types.StatItem{
			Name:  s.Field,
			Count: s.Count,
		})
	}

	// Top应用
	appStats, _ := assetModel.Aggregate(l.ctx, "app", 10)
	topApp := make([]types.StatItem, 0, len(appStats))
	for _, s := range appStats {
		topApp = append(topApp, types.StatItem{
			Name:  s.Field,
			Count: s.Count,
		})
	}

	// Top标题
	titleStats, _ := assetModel.Aggregate(l.ctx, "title", 10)
	topTitle := make([]types.StatItem, 0, len(titleStats))
	for _, s := range titleStats {
		if s.Field != "" {
			topTitle = append(topTitle, types.StatItem{
				Name:  s.Field,
				Count: s.Count,
			})
		}
	}

	// 风险等级分布 
	riskDistribution, _ := assetModel.AggregateRiskLevel(l.ctx)

	return &types.AssetStatResp{
		Code:             0,
		Msg:              "success",
		TotalAsset:       int(totalAsset),
		TotalHost:        int(totalHost),
		NewCount:         int(newCount),
		UpdatedCount:     int(updatedCount),
		TopPorts:         topPorts,
		TopService:       topService,
		TopApp:           topApp,
		TopTitle:         topTitle,
		RiskDistribution: riskDistribution,
	}, nil
}

// AssetDeleteLogic 单个删除
type AssetDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssetDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssetDeleteLogic {
	return &AssetDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssetDeleteLogic) AssetDelete(req *types.AssetDeleteReq, workspaceId string) (resp *types.BaseResp, err error) {
	assetModel := l.svcCtx.GetAssetModel(workspaceId)
	err = assetModel.Delete(l.ctx, req.Id)
	if err != nil {
		return &types.BaseResp{Code: 500, Msg: "删除失败"}, nil
	}
	return &types.BaseResp{Code: 0, Msg: "删除成功"}, nil
}

// AssetBatchDeleteLogic 批量删除
type AssetBatchDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssetBatchDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssetBatchDeleteLogic {
	return &AssetBatchDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssetBatchDeleteLogic) AssetBatchDelete(req *types.AssetBatchDeleteReq, workspaceId string) (resp *types.BaseResp, err error) {
	if len(req.Ids) == 0 {
		return &types.BaseResp{Code: 400, Msg: "请选择要删除的资产"}, nil
	}

	assetModel := l.svcCtx.GetAssetModel(workspaceId)
	deleted, err := assetModel.BatchDelete(l.ctx, req.Ids)
	if err != nil {
		return &types.BaseResp{Code: 500, Msg: "删除失败"}, nil
	}
	return &types.BaseResp{Code: 0, Msg: "成功删除 " + strconv.FormatInt(deleted, 10) + " 条资产"}, nil
}


// AssetHistoryLogic 资产历史记录
type AssetHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssetHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssetHistoryLogic {
	return &AssetHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssetHistoryLogic) AssetHistory(req *types.AssetHistoryReq, workspaceId string) (resp *types.AssetHistoryResp, err error) {
	historyModel := l.svcCtx.GetAssetHistoryModel(workspaceId)

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	histories, err := historyModel.FindByAssetId(l.ctx, req.AssetId, limit)
	if err != nil {
		return &types.AssetHistoryResp{Code: 500, Msg: "查询失败"}, nil
	}

	list := make([]types.AssetHistoryItem, 0, len(histories))
	for _, h := range histories {
		list = append(list, types.AssetHistoryItem{
			Id:         h.Id.Hex(),
			Authority:  h.Authority,
			Host:       h.Host,
			Port:       h.Port,
			Service:    h.Service,
			Title:      h.Title,
			App:        h.App,
			HttpStatus: h.HttpStatus,
			HttpHeader: h.HttpHeader,
			HttpBody:   h.HttpBody,
			Banner:     h.Banner,
			IconHash:   h.IconHash,
			Screenshot: h.Screenshot,
			TaskId:     h.TaskId,
			CreateTime: h.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.AssetHistoryResp{
		Code: 0,
		Msg:  "success",
		List: list,
	}, nil
}
