package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// APIConfig API配置
type APIConfig struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Platform   string             `bson:"platform" json:"platform"` // fofa/hunter/quake
	Key        string             `bson:"key" json:"key"`
	Secret     string             `bson:"secret" json:"secret"`
	Status     string             `bson:"status" json:"status"`
	CreateTime time.Time          `bson:"create_time" json:"createTime"`
	UpdateTime time.Time          `bson:"update_time" json:"updateTime"`
}

// APIConfigModel API配置模型
type APIConfigModel struct {
	coll *mongo.Collection
}

// NewAPIConfigModel 创建API配置模型
func NewAPIConfigModel(db *mongo.Database, workspaceId string) *APIConfigModel {
	return &APIConfigModel{
		coll: db.Collection(workspaceId + "_api_config"),
	}
}

func (m *APIConfigModel) Insert(ctx context.Context, doc *APIConfig) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	_, err := m.coll.InsertOne(ctx, doc)
	return err
}

func (m *APIConfigModel) FindByPlatform(ctx context.Context, platform string) (*APIConfig, error) {
	var doc APIConfig
	err := m.coll.FindOne(ctx, bson.M{"platform": platform, "status": "enable"}).Decode(&doc)
	return &doc, err
}

func (m *APIConfigModel) FindAll(ctx context.Context) ([]APIConfig, error) {
	cursor, err := m.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []APIConfig
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *APIConfigModel) Update(ctx context.Context, id string, update bson.M) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update["update_time"] = time.Now()
	_, err = m.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
	return err
}

func (m *APIConfigModel) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}
