package logic

import (
	"fmt"
	"context"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/model"
	"cscan/onlineapi"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OnlineAPILogic struct {
	ctx context.Context
	svc *svc.ServiceContext
}

func NewOnlineAPILogic(ctx context.Context, svc *svc.ServiceContext) *OnlineAPILogic {
	return &OnlineAPILogic{ctx: ctx, svc: svc}
}

func (l *OnlineAPILogic) Search(req *types.OnlineSearchReq, workspaceId string) (*types.OnlineSearchResp, error) {
	// 获取API配置
	configModel := model.NewAPIConfigModel(l.svc.MongoDB, workspaceId)
	config, err := configModel.FindByPlatform(l.ctx, req.Platform)
	if err != nil {
		return &types.OnlineSearchResp{Code: 404, Msg: "未配置" + req.Platform + "的API密钥"}, nil
	}

	var results []types.OnlineSearchResult
	var total int

	switch req.Platform {
	case "fofa":
		client := onlineapi.NewFofaClient(config.Key, config.Secret)
		result, err := client.Search(l.ctx, req.Query, req.Page, req.PageSize)
		if err != nil {
			return &types.OnlineSearchResp{Code: 500, Msg: "查询失败: " + err.Error()}, nil
		}
		total = result.Size
		assets := client.ParseResults(result)
		for _, a := range assets {
			results = append(results, types.OnlineSearchResult{
				Host: a.Host, IP: a.IP, Port: a.Port, Protocol: a.Protocol,
				Domain: a.Domain, Title: a.Title, Server: a.Server,
				Country: a.Country, City: a.City, Banner: a.Banner,
				ICP: a.ICP, Product: a.Product, OS: a.OS,
			})
		}
	case "hunter":
		client := onlineapi.NewHunterClient(config.Key)
		result, err := client.Search(l.ctx, req.Query, req.Page, req.PageSize, "", "")
		if err != nil {
			return &types.OnlineSearchResp{Code: 500, Msg: "查询失败: " + err.Error()}, nil
		}
		total = result.Data.Total
		for _, a := range result.Data.Arr {
			component := ""
			if len(a.Component) > 0 {
				component = a.Component[0].Name
			}
			results = append(results, types.OnlineSearchResult{
				Host: a.URL, IP: a.IP, Port: a.Port, Protocol: a.Protocol,
				Domain: a.Domain, Title: a.WebTitle, Server: component,
				Country: a.Country, City: a.City, Banner: a.Banner,
				ICP: a.Number, Product: component, OS: a.OS,
			})
		}
	case "quake":
		client := onlineapi.NewQuakeClient(config.Key)
		result, err := client.Search(l.ctx, req.Query, req.Page, req.PageSize)
		if err != nil {
			return &types.OnlineSearchResp{Code: 500, Msg: "查询失败: " + err.Error()}, nil
		}
		total = result.Meta.Pagination.Total
		for _, a := range result.Data {
			results = append(results, types.OnlineSearchResult{
				Host: a.Service.HTTP.Host, IP: a.IP, Port: a.Port, Protocol: a.Service.Name,
				Title: a.Service.HTTP.Title, Server: a.Service.HTTP.Server,
				Country: a.Location.CountryCN, City: a.Location.CityCN,
			})
		}
	default:
		return &types.OnlineSearchResp{Code: 400, Msg: "不支持的平台"}, nil
	}

	return &types.OnlineSearchResp{Code: 0, Msg: "success", Total: total, List: results}, nil
}


func (l *OnlineAPILogic) Import(req *types.OnlineImportReq, workspaceId string) (*types.BaseResp, error) {
	assetModel := l.svc.GetAssetModel(workspaceId)

	count := 0
	for _, a := range req.Assets {
		asset := &model.Asset{
			Authority: a.Host,
			Host:      a.IP,
			Port:      a.Port,
			Service:   a.Protocol,
			Title:     a.Title,
			App:       []string{a.Product},
			Source:    "onlineapi",
		}
		if err := assetModel.Upsert(l.ctx, asset); err == nil {
			count++
		}
	}

	return &types.BaseResp{Code: 0, Msg: fmt.Sprintf("成功导入%d条资产", count)}, nil
}

func (l *OnlineAPILogic) ConfigList(workspaceId string) (*types.APIConfigListResp, error) {
	configModel := model.NewAPIConfigModel(l.svc.MongoDB, workspaceId)
	docs, err := configModel.FindAll(l.ctx)
	if err != nil {
		return &types.APIConfigListResp{Code: 500, Msg: "查询失败"}, nil
	}

	list := make([]types.APIConfig, 0, len(docs))
	for _, doc := range docs {
		list = append(list, types.APIConfig{
			Id:         doc.Id.Hex(),
			Platform:   doc.Platform,
			Key:        doc.Key,
			Secret:     maskSecret(doc.Secret),
			Status:     doc.Status,
			CreateTime: doc.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.APIConfigListResp{Code: 0, Msg: "success", List: list}, nil
}

func (l *OnlineAPILogic) ConfigSave(req *types.APIConfigSaveReq, workspaceId string) (*types.BaseResp, error) {
	configModel := model.NewAPIConfigModel(l.svc.MongoDB, workspaceId)

	if req.Id != "" {
		update := bson.M{
			"key":         req.Key,
			"secret":      req.Secret,
			"update_time": time.Now(),
		}
		if err := configModel.Update(l.ctx, req.Id, update); err != nil {
			return &types.BaseResp{Code: 500, Msg: "更新失败"}, nil
		}
	} else {
		doc := &model.APIConfig{
			Id:       primitive.NewObjectID(),
			Platform: req.Platform,
			Key:      req.Key,
			Secret:   req.Secret,
			Status:   "enable",
		}
		if err := configModel.Insert(l.ctx, doc); err != nil {
			return &types.BaseResp{Code: 500, Msg: "保存失败"}, nil
		}
	}

	return &types.BaseResp{Code: 0, Msg: "保存成功"}, nil
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
