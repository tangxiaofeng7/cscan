package svc

import (
	"context"
	"time"

	"cscan/model"
	"cscan/rpc/task/internal/config"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ServiceContext struct {
	Config                  config.Config
	MongoClient             *mongo.Client
	MongoDB                 *mongo.Database
	RedisClient             *redis.Client
	NucleiTemplateModel     *model.NucleiTemplateModel
	FingerprintModel        *model.FingerprintModel
	CustomPocModel          *model.CustomPocModel
	HttpServiceMappingModel *model.HttpServiceMappingModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(c.Mongo.Uri))
	if err != nil {
		panic(err)
	}

	mongoDB := mongoClient.Database(c.Mongo.DbName)

	// 使用go-zero Redis配置
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.RedisConf.Host,
		Password: c.RedisConf.Pass,
		DB:       0,
	})

	return &ServiceContext{
		Config:                  c,
		MongoClient:             mongoClient,
		MongoDB:                 mongoDB,
		RedisClient:             rdb,
		NucleiTemplateModel:     model.NewNucleiTemplateModel(mongoDB),
		FingerprintModel:        model.NewFingerprintModel(mongoDB),
		CustomPocModel:          model.NewCustomPocModel(mongoDB),
		HttpServiceMappingModel: model.NewHttpServiceMappingModel(mongoDB),
	}
}

func (s *ServiceContext) GetAssetModel(workspaceId string) *model.AssetModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewAssetModel(s.MongoDB, workspaceId)
}

func (s *ServiceContext) GetMainTaskModel(workspaceId string) *model.MainTaskModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewMainTaskModel(s.MongoDB, workspaceId)
}

func (s *ServiceContext) GetVulModel(workspaceId string) *model.VulModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewVulModel(s.MongoDB, workspaceId)
}

func (s *ServiceContext) GetExecutorTaskModel(workspaceId string) *model.ExecutorTaskModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewExecutorTaskModel(s.MongoDB, workspaceId)
}

func (s *ServiceContext) GetAssetHistoryModel(workspaceId string) *model.AssetHistoryModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewAssetHistoryModel(s.MongoDB, workspaceId)
}
