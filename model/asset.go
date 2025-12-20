package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IPV4 struct {
	IPName   string `bson:"ip" json:"ip"`
	IPInt    uint32 `bson:"uint32" json:"uint32"`
	Location string `bson:"location" json:"location"`
}

type IPV6 struct {
	IPName   string `bson:"ip" json:"ip"`
	Location string `bson:"location" json:"location"`
}

type IP struct {
	IpV4 []IPV4 `bson:"ipv4,omitempty" json:"ipv4,omitempty"`
	IpV6 []IPV6 `bson:"ipv6,omitempty" json:"ipv6,omitempty"`
}

type Asset struct {
	Id            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Authority     string             `bson:"authority" json:"authority"`
	Host          string             `bson:"host" json:"host"`
	Port          int                `bson:"port" json:"port"`
	Category      string             `bson:"category" json:"category"`
	Ip            IP                 `bson:"ip" json:"ip"`
	Domain        string             `bson:"domain,omitempty" json:"domain"`
	Service       string             `bson:"service,omitempty" json:"service"`
	Server        string             `bson:"server,omitempty" json:"server"`
	Banner        string             `bson:"banner,omitempty" json:"banner"`
	Title         string             `bson:"title,omitempty" json:"title"`
	App           []string           `bson:"app,omitempty" json:"app"`
	HttpStatus    string             `bson:"status,omitempty" json:"httpStatus"`
	HttpHeader    string             `bson:"header,omitempty" json:"httpHeader"`
	HttpBody      string             `bson:"body,omitempty" json:"httpBody"`
	Cert          string             `bson:"cert,omitempty" json:"cert"`
	IconHash      string             `bson:"icon_hash,omitempty" json:"iconHash"`
	IconHashFile  string             `bson:"icon_hash_file,omitempty" json:"iconHashFile"`
	IconHashBytes []byte             `bson:"icon_hash_bytes,omitempty" json:"-"`
	Screenshot    string             `bson:"screenshot,omitempty" json:"screenshot"`
	OrgId         string             `bson:"org,omitempty" json:"orgId"`
	ColorTag      string             `bson:"color,omitempty" json:"colorTag"`
	Memo          string             `bson:"memo,omitempty" json:"memo"`
	IsCDN         bool               `bson:"cdn,omitempty" json:"isCdn"`
	CName         string             `bson:"cname,omitempty" json:"cname"`
	IsCloud       bool               `bson:"cloud,omitempty" json:"isCloud"`
	IsNewAsset    bool               `bson:"new" json:"isNew"`
	IsUpdated     bool               `bson:"update" json:"isUpdated"`
	TaskId        string             `bson:"taskId" json:"taskId"`
	Source        string             `bson:"source,omitempty" json:"source"`
	CreateTime    time.Time          `bson:"create_time" json:"createTime"`
	UpdateTime    time.Time          `bson:"update_time" json:"updateTime"`
}

type AssetModel struct {
	coll *mongo.Collection
}

func NewAssetModel(db *mongo.Database, workspaceId string) *AssetModel {
	return &AssetModel{
		coll: db.Collection(workspaceId + "_asset"),
	}
}

func (m *AssetModel) Insert(ctx context.Context, doc *Asset) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	doc.IsNewAsset = true
	_, err := m.coll.InsertOne(ctx, doc)
	return err
}

func (m *AssetModel) FindById(ctx context.Context, id string) (*Asset, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var doc Asset
	err = m.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	return &doc, err
}

func (m *AssetModel) FindByAuthority(ctx context.Context, authority, taskId string) (*Asset, error) {
	var doc Asset
	filter := bson.M{"authority": authority, "taskId": taskId}
	err := m.coll.FindOne(ctx, filter).Decode(&doc)
	return &doc, err
}

func (m *AssetModel) FindByHostPort(ctx context.Context, host string, port int) (*Asset, error) {
	var doc Asset
	filter := bson.M{"host": host, "port": port}
	err := m.coll.FindOne(ctx, filter).Decode(&doc)
	return &doc, err
}

func (m *AssetModel) Find(ctx context.Context, filter bson.M, page, pageSize int) ([]Asset, error) {
	return m.FindWithSort(ctx, filter, page, pageSize, "update_time")
}

func (m *AssetModel) FindWithSort(ctx context.Context, filter bson.M, page, pageSize int, sortField string) ([]Asset, error) {
	opts := options.Find()
	if page > 0 && pageSize > 0 {
		opts.SetSkip(int64((page - 1) * pageSize))
		opts.SetLimit(int64(pageSize))
	}
	opts.SetSort(bson.D{{Key: sortField, Value: -1}})

	cursor, err := m.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []Asset
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *AssetModel) Count(ctx context.Context, filter bson.M) (int64, error) {
	return m.coll.CountDocuments(ctx, filter)
}

func (m *AssetModel) Update(ctx context.Context, id string, update bson.M) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update["update_time"] = time.Now()
	_, err = m.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
	return err
}

func (m *AssetModel) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

func (m *AssetModel) BatchDelete(ctx context.Context, ids []string) (int64, error) {
	oids := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		oids = append(oids, oid)
	}
	if len(oids) == 0 {
		return 0, nil
	}
	result, err := m.coll.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": oids}})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (m *AssetModel) Aggregate(ctx context.Context, field string, limit int) ([]StatResult, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$" + field},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := m.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []StatResult
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

type StatResult struct {
	Field string `bson:"_id"`
	Count int    `bson:"count"`
}

// AggregatePort 专门用于端口统计（端口是int类型）
func (m *AssetModel) AggregatePort(ctx context.Context, limit int) ([]PortStatResult, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$port"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := m.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []PortStatResult
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

type PortStatResult struct {
	Port  int `bson:"_id"`
	Count int `bson:"count"`
}

// AssetHistory 资产历史记录
type AssetHistory struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AssetId    string             `bson:"assetId" json:"assetId"`
	Authority  string             `bson:"authority" json:"authority"`
	Host       string             `bson:"host" json:"host"`
	Port       int                `bson:"port" json:"port"`
	Service    string             `bson:"service,omitempty" json:"service"`
	Title      string             `bson:"title,omitempty" json:"title"`
	App        []string           `bson:"app,omitempty" json:"app"`
	HttpStatus string             `bson:"status,omitempty" json:"httpStatus"`
	HttpHeader string             `bson:"header,omitempty" json:"httpHeader"`
	HttpBody   string             `bson:"body,omitempty" json:"httpBody"`
	Banner     string             `bson:"banner,omitempty" json:"banner"`
	IconHash   string             `bson:"icon_hash,omitempty" json:"iconHash"`
	Screenshot string             `bson:"screenshot,omitempty" json:"screenshot"`
	TaskId     string             `bson:"taskId" json:"taskId"`
	CreateTime time.Time          `bson:"create_time" json:"createTime"`
}

// AssetHistoryModel 资产历史模型
type AssetHistoryModel struct {
	coll *mongo.Collection
}

func NewAssetHistoryModel(db *mongo.Database, workspaceId string) *AssetHistoryModel {
	return &AssetHistoryModel{
		coll: db.Collection(workspaceId + "_asset_history"),
	}
}

func (m *AssetHistoryModel) Insert(ctx context.Context, doc *AssetHistory) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	doc.CreateTime = time.Now()
	_, err := m.coll.InsertOne(ctx, doc)
	return err
}

func (m *AssetHistoryModel) FindByAssetId(ctx context.Context, assetId string, limit int) ([]AssetHistory, error) {
	opts := options.Find()
	opts.SetSort(bson.D{{Key: "create_time", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := m.coll.Find(ctx, bson.M{"assetId": assetId}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []AssetHistory
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *AssetHistoryModel) FindByAuthority(ctx context.Context, authority string, limit int) ([]AssetHistory, error) {
	opts := options.Find()
	opts.SetSort(bson.D{{Key: "create_time", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := m.coll.Find(ctx, bson.M{"authority": authority}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []AssetHistory
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// Upsert 插入或更新资产
func (m *AssetModel) Upsert(ctx context.Context, doc *Asset) error {
	filter := bson.M{"authority": doc.Authority}
	if doc.TaskId != "" {
		filter["taskId"] = doc.TaskId
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"host":        doc.Host,
			"port":        doc.Port,
			"service":     doc.Service,
			"title":       doc.Title,
			"app":         doc.App,
			"source":      doc.Source,
			"update_time": now,
		},
		"$setOnInsert": bson.M{
			"_id":         primitive.NewObjectID(),
			"create_time": now,
			"new":         true,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := m.coll.UpdateOne(ctx, filter, update, opts)
	return err
}
