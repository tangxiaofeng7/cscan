package logic

import (
	"context"
	"time"

	"cscan/model"
	"cscan/pkg/utils"
	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SaveTaskResultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveTaskResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveTaskResultLogic {
	return &SaveTaskResultLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SaveTaskResult 保存任务结果
func (l *SaveTaskResultLogic) SaveTaskResult(in *pb.SaveTaskResultReq) (*pb.SaveTaskResultResp, error) {
	if len(in.Assets) == 0 {
		return &pb.SaveTaskResultResp{
			Success: true,
			Message: "No assets to save",
		}, nil
	}

	workspaceId := in.WorkspaceId
	if workspaceId == "" {
		workspaceId = "default"
	}

	assetModel := l.svcCtx.GetAssetModel(workspaceId)

	var totalAsset, newAsset, updateAsset int32
	now := time.Now()

	for _, pbAsset := range in.Assets {
		// 转换为model.Asset
		asset := &model.Asset{
			Authority:     pbAsset.Authority,
			Host:          pbAsset.Host,
			Port:          int(pbAsset.Port),
			Category:      pbAsset.Category,
			Service:       pbAsset.Service,
			Title:         pbAsset.Title,
			App:           pbAsset.App,
			HttpStatus:    pbAsset.HttpStatus,
			HttpHeader:    pbAsset.HttpHeader,
			HttpBody:      pbAsset.HttpBody,
			IconHash:      pbAsset.IconHash,
			IconHashBytes: pbAsset.IconData,
			Screenshot:    pbAsset.Screenshot,
			Server:        pbAsset.Server,
			Banner:        pbAsset.Banner,
			IsHTTP:        pbAsset.IsHttp,
			TaskId:        in.MainTaskId,
			Source:        pbAsset.Source,
			OrgId:         in.OrgId,
		}

		// 如果Source为空，设置默认值
		if asset.Source == "" {
			asset.Source = "scan"
		}

		// 处理IP信息
		if len(pbAsset.Ipv4) > 0 {
			for _, ip := range pbAsset.Ipv4 {
				asset.Ip.IpV4 = append(asset.Ip.IpV4, model.IPV4{
					IPName:   ip.Ip,
					Location: ip.Location,
				})
			}
		}
		if len(pbAsset.Ipv6) > 0 {
			for _, ip := range pbAsset.Ipv6 {
				asset.Ip.IpV6 = append(asset.Ip.IpV6, model.IPV6{
					IPName:   ip.Ip,
					Location: ip.Location,
				})
			}
		}

		// 处理CName
		if pbAsset.Cname != "" {
			asset.CName = pbAsset.Cname
		}

		// 设置Domain字段 - 如果Host不是IP地址，则设置为Domain
		if asset.Category == "domain" || !utils.IsIPAddress(asset.Host) {
			asset.Domain = asset.Host
		}

		// 检查是否已存在
		var existing *model.Asset
		var err error

		if asset.Port > 0 {
			// 有端口的资产，按host:port查找
			existing, err = assetModel.FindByHostPort(l.ctx, asset.Host, asset.Port)
		} else {
			// 无端口的资产（如域名），按authority查找（不限制taskId）
			existing, err = assetModel.FindByAuthorityOnly(l.ctx, asset.Authority)
		}

		if err != nil || existing == nil {
			// 新资产
			asset.Id = primitive.NewObjectID()
			asset.CreateTime = now
			asset.UpdateTime = now
			asset.IsNewAsset = true
			asset.IsUpdated = false

			if err := assetModel.Insert(l.ctx, asset); err != nil {
				l.Logger.Errorf("Insert asset failed: %v", err)
				continue
			}
			newAsset++
		} else {
			// 更新已存在的资产
			// 只有当任务ID不同时才保存历史记录（表示是新一轮扫描，需要记录上一次的状态）
			if existing.TaskId != "" && existing.TaskId != in.MainTaskId {
				historyModel := l.svcCtx.GetAssetHistoryModel(workspaceId)
				
				// 检查是否已存在同一任务的历史记录（避免重复）
				exists, _ := historyModel.ExistsByAssetIdAndTaskId(l.ctx, existing.Id.Hex(), existing.TaskId)
				if !exists {
					// 保存上一次扫描的状态作为历史记录
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
						IconHash:   existing.IconHash,
						Screenshot: existing.Screenshot,
						Banner:     existing.Banner,
						TaskId:     existing.TaskId, // 使用旧的任务ID
						CreateTime: existing.UpdateTime, // 使用旧的更新时间
					}
					if err := historyModel.Insert(l.ctx, history); err != nil {
						l.Logger.Errorf("Insert asset history failed: %v", err)
						// 继续更新资产，不中断
					}
				}
			}

			// 更新资产
			updateFields := map[string]interface{}{
				"authority":   asset.Authority,
				"service":     asset.Service,
				"title":       asset.Title,
				"app":         asset.App,
				"status":      asset.HttpStatus,
				"header":      asset.HttpHeader,
				"body":        asset.HttpBody,
				"icon_hash":   asset.IconHash,
				"screenshot":  asset.Screenshot,
				"server":      asset.Server,
				"banner":      asset.Banner,
				"is_http":     asset.IsHTTP,
				"taskId":      asset.TaskId,
				"update_time": now,
				"update":      true,
				"new":         false,
			}

			// 更新 IconData
			if len(asset.IconHashBytes) > 0 {
				updateFields["icon_hash_bytes"] = asset.IconHashBytes
			}

			// 更新IP信息
			if len(asset.Ip.IpV4) > 0 || len(asset.Ip.IpV6) > 0 {
				updateFields["ip"] = asset.Ip
			}

			// 更新CName
			if asset.CName != "" {
				updateFields["cname"] = asset.CName
			}

			// 更新Domain
			if asset.Domain != "" {
				updateFields["domain"] = asset.Domain
			}

			// 更新OrgId
			if asset.OrgId != "" {
				updateFields["org_id"] = asset.OrgId
			}

			// 更新Source
			if asset.Source != "" {
				updateFields["source"] = asset.Source
			}

			// 更新Category
			if asset.Category != "" {
				updateFields["category"] = asset.Category
			}

			if err := assetModel.Update(l.ctx, existing.Id.Hex(), updateFields); err != nil {
				l.Logger.Errorf("Update asset failed: %v", err)
				continue
			}
			updateAsset++
		}
		totalAsset++
	}

	l.Logger.Infof("SaveTaskResult: total=%d, new=%d, update=%d", totalAsset, newAsset, updateAsset)

	return &pb.SaveTaskResultResp{
		Success:     true,
		Message:     "Assets saved successfully",
		TotalAsset:  totalAsset,
		NewAsset:    newAsset,
		UpdateAsset: updateAsset,
	}, nil
}
