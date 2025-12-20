package model

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Fingerprint 指纹规则
type Fingerprint struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`               // 应用名称
	Category    string             `bson:"category" json:"category"`       // 分类: cms, framework, server, etc.
	Website     string             `bson:"website" json:"website"`         // 官网
	Icon        string             `bson:"icon" json:"icon"`               // 图标URL
	Description string             `bson:"description" json:"description"` // 描述
	// 匹配规则 - Wappalyzer格式
	Headers   map[string]string `bson:"headers" json:"headers"`     // HTTP头匹配 {"Server": "nginx"}
	Cookies   map[string]string `bson:"cookies" json:"cookies"`     // Cookie匹配
	HTML      []string          `bson:"html" json:"html"`           // HTML内容匹配（正则）
	Scripts   []string          `bson:"scripts" json:"scripts"`     // JS脚本路径匹配（正则）
	ScriptSrc []string          `bson:"scriptSrc" json:"scriptSrc"` // Script src匹配
	JS        map[string]string `bson:"js" json:"js"`               // JS变量匹配
	Meta      map[string]string `bson:"meta" json:"meta"`           // Meta标签匹配
	CSS       []string          `bson:"css" json:"css"`             // CSS匹配（正则）
	URL       []string          `bson:"url" json:"url"`             // URL路径匹配（正则）
	Dom       string            `bson:"dom" json:"dom"`             // DOM选择器匹配（JSON字符串）
	// 匹配规则 - ARL/自定义格式（简化规则语法）
	Rule      string            `bson:"rule" json:"rule"`           // ARL格式规则: body="xxx" && title="xxx"
	// 其他
	Implies    []string  `bson:"implies" json:"implies"`       // 隐含的其他技术
	Excludes   []string  `bson:"excludes" json:"excludes"`     // 排除的技术
	CPE        string    `bson:"cpe" json:"cpe"`               // CPE标识
	Source     string    `bson:"source" json:"source"`         // 来源: wappalyzer, arl, custom
	IsBuiltin  bool      `bson:"is_builtin" json:"isBuiltin"`  // 是否内置指纹
	Enabled    bool      `bson:"enabled" json:"enabled"`       // 是否启用
	CreateTime time.Time `bson:"create_time" json:"createTime"`
	UpdateTime time.Time `bson:"update_time" json:"updateTime"`
}

// FingerprintModel 指纹模型
type FingerprintModel struct {
	coll *mongo.Collection
}

func NewFingerprintModel(db *mongo.Database) *FingerprintModel {
	coll := db.Collection("fingerprint")
	// 创建索引 - name+rule 组合唯一，允许同名不同规则的指纹
	coll.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}, {Key: "rule", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "name", Value: 1}}},
		{Keys: bson.D{{Key: "category", Value: 1}}},
		{Keys: bson.D{{Key: "is_builtin", Value: 1}}},
		{Keys: bson.D{{Key: "enabled", Value: 1}}},
	})
	return &FingerprintModel{coll: coll}
}

func (m *FingerprintModel) Insert(ctx context.Context, doc *Fingerprint) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	_, err := m.coll.InsertOne(ctx, doc)
	return err
}

func (m *FingerprintModel) Upsert(ctx context.Context, doc *Fingerprint) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	doc.UpdateTime = time.Now()
	if doc.CreateTime.IsZero() {
		doc.CreateTime = doc.UpdateTime
	}

	filter := bson.M{"name": doc.Name}
	update := bson.M{"$set": doc}
	opts := options.Update().SetUpsert(true)
	_, err := m.coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func (m *FingerprintModel) Find(ctx context.Context, filter bson.M, page, pageSize int) ([]Fingerprint, error) {
	opts := options.Find()
	if page > 0 && pageSize > 0 {
		opts.SetSkip(int64((page - 1) * pageSize))
		opts.SetLimit(int64(pageSize))
	}
	opts.SetSort(bson.D{{Key: "category", Value: 1}, {Key: "name", Value: 1}})

	cursor, err := m.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []Fingerprint
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *FingerprintModel) FindAll(ctx context.Context) ([]Fingerprint, error) {
	return m.Find(ctx, bson.M{}, 0, 0)
}

func (m *FingerprintModel) FindEnabled(ctx context.Context) ([]Fingerprint, error) {
	return m.Find(ctx, bson.M{"enabled": true}, 0, 0)
}

func (m *FingerprintModel) FindCustom(ctx context.Context) ([]Fingerprint, error) {
	return m.Find(ctx, bson.M{"is_builtin": false}, 0, 0)
}

func (m *FingerprintModel) FindBuiltin(ctx context.Context) ([]Fingerprint, error) {
	return m.Find(ctx, bson.M{"is_builtin": true}, 0, 0)
}

func (m *FingerprintModel) FindById(ctx context.Context, id string) (*Fingerprint, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var doc Fingerprint
	err = m.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	return &doc, err
}

func (m *FingerprintModel) FindByName(ctx context.Context, name string) (*Fingerprint, error) {
	var doc Fingerprint
	err := m.coll.FindOne(ctx, bson.M{"name": name}).Decode(&doc)
	return &doc, err
}

func (m *FingerprintModel) Count(ctx context.Context, filter bson.M) (int64, error) {
	return m.coll.CountDocuments(ctx, filter)
}

func (m *FingerprintModel) Update(ctx context.Context, id string, update bson.M) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update["update_time"] = time.Now()
	_, err = m.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
	return err
}

func (m *FingerprintModel) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

func (m *FingerprintModel) GetCategories(ctx context.Context) ([]string, error) {
	results, err := m.coll.Distinct(ctx, "category", bson.M{})
	if err != nil {
		return nil, err
	}
	categories := make([]string, 0, len(results))
	for _, r := range results {
		if s, ok := r.(string); ok && s != "" {
			categories = append(categories, s)
		}
	}
	return categories, nil
}

func (m *FingerprintModel) GetStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// 总数
	total, _ := m.coll.CountDocuments(ctx, bson.M{})
	stats["total"] = total

	// 内置数量
	builtin, _ := m.coll.CountDocuments(ctx, bson.M{"is_builtin": true})
	stats["builtin"] = builtin

	// 自定义数量
	custom, _ := m.coll.CountDocuments(ctx, bson.M{"is_builtin": false})
	stats["custom"] = custom

	// 启用数量
	enabled, _ := m.coll.CountDocuments(ctx, bson.M{"enabled": true})
	stats["enabled"] = enabled

	return stats, nil
}

// DeleteAll 删除所有指纹
func (m *FingerprintModel) DeleteAll(ctx context.Context) error {
	_, err := m.coll.DeleteMany(ctx, bson.M{})
	return err
}

// DeleteBuiltin 删除所有内置指纹
func (m *FingerprintModel) DeleteBuiltin(ctx context.Context) error {
	_, err := m.coll.DeleteMany(ctx, bson.M{"is_builtin": true})
	return err
}

// DeleteCustom 删除所有自定义指纹（非内置）
func (m *FingerprintModel) DeleteCustom(ctx context.Context) (int64, error) {
	result, err := m.coll.DeleteMany(ctx, bson.M{"is_builtin": false})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// DeleteBySource 按来源删除指纹
func (m *FingerprintModel) DeleteBySource(ctx context.Context, source string) (int64, error) {
	result, err := m.coll.DeleteMany(ctx, bson.M{"source": source})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}


// BulkUpsert 批量插入或更新指纹
// 去重原则：只有 name 和 rule 都完全相同才视为重复
// 返回: 新插入数量, 更新数量(包括匹配但未修改的), 错误
func (m *FingerprintModel) BulkUpsert(ctx context.Context, docs []*Fingerprint) (int, int, error) {
	if len(docs) == 0 {
		return 0, 0, nil
	}

	var models []mongo.WriteModel
	now := time.Now()

	for _, doc := range docs {
		if doc.Id.IsZero() {
			doc.Id = primitive.NewObjectID()
		}
		doc.UpdateTime = now
		if doc.CreateTime.IsZero() {
			doc.CreateTime = now
		}

		// 使用 name + rule 作为去重条件，只有两者都相同才视为重复
		filter := bson.M{"name": doc.Name, "rule": doc.Rule}
		update := bson.M{"$set": doc}
		model := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true)
		models = append(models, model)
	}

	// 分批执行，每批500条
	batchSize := 500
	var inserted, matched int

	for i := 0; i < len(models); i += batchSize {
		end := i + batchSize
		if end > len(models) {
			end = len(models)
		}

		result, err := m.coll.BulkWrite(ctx, models[i:end], options.BulkWrite().SetOrdered(false))
		if err != nil {
			// 记录错误但继续处理
			fmt.Printf("BulkWrite error: %v\n", err)
			continue
		}
		inserted += int(result.UpsertedCount)
		// MatchedCount 包括已存在的记录（无论是否修改）
		matched += int(result.MatchedCount)
	}

	// 返回新插入数量和匹配更新数量
	return inserted, matched, nil
}
