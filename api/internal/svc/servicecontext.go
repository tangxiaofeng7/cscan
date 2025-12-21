package svc

import (
	"context"
	"time"

	"cscan/api/internal/config"
	"cscan/api/internal/svc/sync"
	"cscan/model"
	"cscan/rpc/task/pb"
	"cscan/scheduler"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ServiceContext struct {
	Config                  config.Config
	MongoClient             *mongo.Client
	MongoDB                 *mongo.Database
	RedisClient             *redis.Client
	TaskRpcClient           pb.TaskServiceClient
	UserModel               *model.UserModel
	WorkspaceModel          *model.WorkspaceModel
	ProfileModel            *model.TaskProfileModel
	TagMappingModel         *model.TagMappingModel
	CustomPocModel          *model.CustomPocModel
	NucleiTemplateModel     *model.NucleiTemplateModel
	FingerprintModel        *model.FingerprintModel
	HttpServiceMappingModel *model.HttpServiceMappingModel

	// 调度器
	Scheduler *scheduler.Scheduler

	// 同步服务
	SyncMethods *sync.SyncMethods

	// 缓存的模板元数据
	TemplateCategories []string
	TemplateTags       []string
	TemplateStats      map[string]int
}

func NewServiceContext(c config.Config) *ServiceContext {
	// MongoDB连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(c.Mongo.Uri))
	if err != nil {
		panic(err)
	}

	mongoDB := mongoClient.Database(c.Mongo.DbName)

	// Redis连接 - 使用go-zero配置
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Host,
		Password: c.Redis.Pass,
		DB:       0,
	})

	// 创建RPC客户端
	taskRpcClient := pb.NewTaskServiceClient(zrpc.MustNewClient(c.TaskRpc).Conn())

	svcCtx := &ServiceContext{
		Config:                  c,
		MongoClient:             mongoClient,
		MongoDB:                 mongoDB,
		RedisClient:             rdb,
		TaskRpcClient:           taskRpcClient,
		UserModel:               model.NewUserModel(mongoDB),
		WorkspaceModel:          model.NewWorkspaceModel(mongoDB),
		ProfileModel:            model.NewTaskProfileModel(mongoDB),
		TagMappingModel:         model.NewTagMappingModel(mongoDB),
		CustomPocModel:          model.NewCustomPocModel(mongoDB),
		NucleiTemplateModel:     model.NewNucleiTemplateModel(mongoDB),
		FingerprintModel:        model.NewFingerprintModel(mongoDB),
		HttpServiceMappingModel: model.NewHttpServiceMappingModel(mongoDB),
		Scheduler:               scheduler.NewScheduler(rdb),
		TemplateCategories:      []string{},
		TemplateTags:            []string{},
		TemplateStats:           map[string]int{},
	}

	// 初始化同步服务
	svcCtx.SyncMethods = sync.NewSyncMethods(
		svcCtx.NucleiTemplateModel,
		svcCtx.FingerprintModel,
		svcCtx.CustomPocModel,
	)

	return svcCtx
}

// GetAssetModel 根据workspaceId获取资产模型
func (s *ServiceContext) GetAssetModel(workspaceId string) *model.AssetModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewAssetModel(s.MongoDB, workspaceId)
}

// GetMainTaskModel 根据workspaceId获取主任务模型
func (s *ServiceContext) GetMainTaskModel(workspaceId string) *model.MainTaskModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewMainTaskModel(s.MongoDB, workspaceId)
}

// GetVulModel 根据workspaceId获取漏洞模型
func (s *ServiceContext) GetVulModel(workspaceId string) *model.VulModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewVulModel(s.MongoDB, workspaceId)
}

// GetAssetHistoryModel 根据workspaceId获取资产历史模型
func (s *ServiceContext) GetAssetHistoryModel(workspaceId string) *model.AssetHistoryModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewAssetHistoryModel(s.MongoDB, workspaceId)
}

// RefreshTemplateCache 刷新模板元数据缓存
func (s *ServiceContext) RefreshTemplateCache() {
	ctx := context.Background()

	categories, err := s.NucleiTemplateModel.GetCategories(ctx)
	if err == nil {
		s.TemplateCategories = categories
	}

	s.TemplateTags = []string{}

	stats, err := s.NucleiTemplateModel.GetStats(ctx)
	if err == nil {
		s.TemplateStats = stats
	}

	logx.Infof("[NucleiCache] Refreshed: %d categories, stats: %v", len(s.TemplateCategories), s.TemplateStats)
}


// SyncNucleiTemplates 同步Nuclei模板
func (s *ServiceContext) SyncNucleiTemplates() {
	s.SyncMethods.SyncNucleiTemplates()
}

// SyncWappalyzerFingerprints 同步Wappalyzer指纹
func (s *ServiceContext) SyncWappalyzerFingerprints() {
	s.SyncMethods.SyncWappalyzerFingerprints()
}

// ImportCustomPocAndFingerprints 导入自定义POC和指纹
func (s *ServiceContext) ImportCustomPocAndFingerprints() {
	s.SyncMethods.ImportCustomPocAndFingerprints()
}
