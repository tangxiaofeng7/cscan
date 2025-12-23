package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TagMapping 应用标签到Nuclei标签的映射
// 用于基于Wappalyzer识别的应用自动选择对应的POC
type TagMapping struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AppName     string             `bson:"app_name" json:"appName"`         // Wappalyzer识别的应用名称
	NucleiTags  []string           `bson:"nuclei_tags" json:"nucleiTags"`   // 对应的Nuclei标签
	Description string             `bson:"description" json:"description"`  // 描述
	Enabled     bool               `bson:"enabled" json:"enabled"`          // 是否启用
	CreateTime  time.Time          `bson:"create_time" json:"createTime"`
	UpdateTime  time.Time          `bson:"update_time" json:"updateTime"`
}

// CustomPoc 自定义POC
type CustomPoc struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`               // POC名称
	TemplateId  string             `bson:"template_id" json:"templateId"`  // 模板ID（唯一标识）
	Severity    string             `bson:"severity" json:"severity"`       // 严重级别: critical/high/medium/low/info
	Tags        []string           `bson:"tags" json:"tags"`               // 标签
	Author      string             `bson:"author" json:"author"`           // 作者
	Description string             `bson:"description" json:"description"` // 描述
	Content     string             `bson:"content" json:"content"`         // YAML内容
	Enabled     bool               `bson:"enabled" json:"enabled"`         // 是否启用
	CreateTime  time.Time          `bson:"create_time" json:"createTime"`
	UpdateTime  time.Time          `bson:"update_time" json:"updateTime"`
}

// TagMappingModel 标签映射模型
type TagMappingModel struct {
	coll *mongo.Collection
}

func NewTagMappingModel(db *mongo.Database) *TagMappingModel {
	return &TagMappingModel{
		coll: db.Collection("tag_mapping"),
	}
}

func (m *TagMappingModel) Insert(ctx context.Context, doc *TagMapping) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	_, err := m.coll.InsertOne(ctx, doc)
	return err
}

func (m *TagMappingModel) FindAll(ctx context.Context) ([]TagMapping, error) {
	opts := options.Find().SetSort(bson.D{{Key: "app_name", Value: 1}})
	cursor, err := m.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []TagMapping
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *TagMappingModel) FindEnabled(ctx context.Context) ([]TagMapping, error) {
	cursor, err := m.coll.Find(ctx, bson.M{"enabled": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []TagMapping
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *TagMappingModel) FindById(ctx context.Context, id string) (*TagMapping, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var doc TagMapping
	err = m.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	return &doc, err
}

func (m *TagMappingModel) FindByAppName(ctx context.Context, appName string) (*TagMapping, error) {
	var doc TagMapping
	err := m.coll.FindOne(ctx, bson.M{"app_name": appName}).Decode(&doc)
	return &doc, err
}

func (m *TagMappingModel) Update(ctx context.Context, id string, doc *TagMapping) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update := bson.M{
		"app_name":    doc.AppName,
		"nuclei_tags": doc.NucleiTags,
		"description": doc.Description,
		"enabled":     doc.Enabled,
		"update_time": time.Now(),
	}
	_, err = m.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
	return err
}

func (m *TagMappingModel) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

// CustomPocModel 自定义POC模型
type CustomPocModel struct {
	coll *mongo.Collection
}

func NewCustomPocModel(db *mongo.Database) *CustomPocModel {
	return &CustomPocModel{
		coll: db.Collection("custom_poc"),
	}
}

func (m *CustomPocModel) Insert(ctx context.Context, doc *CustomPoc) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	_, err := m.coll.InsertOne(ctx, doc)
	return err
}

func (m *CustomPocModel) FindAll(ctx context.Context, page, pageSize int) ([]CustomPoc, error) {
	opts := options.Find()
	if page > 0 && pageSize > 0 {
		opts.SetSkip(int64((page - 1) * pageSize))
		opts.SetLimit(int64(pageSize))
	}
	opts.SetSort(bson.D{{Key: "create_time", Value: -1}})

	cursor, err := m.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []CustomPoc
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// FindWithFilter 带筛选条件的查询
func (m *CustomPocModel) FindWithFilter(ctx context.Context, filter bson.M, page, pageSize int) ([]CustomPoc, error) {
	opts := options.Find()
	if page > 0 && pageSize > 0 {
		opts.SetSkip(int64((page - 1) * pageSize))
		opts.SetLimit(int64(pageSize))
	}
	opts.SetSort(bson.D{{Key: "create_time", Value: -1}})

	cursor, err := m.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []CustomPoc
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// CountWithFilter 带筛选条件的计数
func (m *CustomPocModel) CountWithFilter(ctx context.Context, filter bson.M) (int64, error) {
	return m.coll.CountDocuments(ctx, filter)
}

func (m *CustomPocModel) FindEnabled(ctx context.Context) ([]CustomPoc, error) {
	cursor, err := m.coll.Find(ctx, bson.M{"enabled": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []CustomPoc
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *CustomPocModel) FindByTags(ctx context.Context, tags []string) ([]CustomPoc, error) {
	cursor, err := m.coll.Find(ctx, bson.M{
		"enabled": true,
		"tags":    bson.M{"$in": tags},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []CustomPoc
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *CustomPocModel) FindById(ctx context.Context, id string) (*CustomPoc, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var doc CustomPoc
	err = m.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	return &doc, err
}

// FindByTemplateId 根据模板ID查找自定义POC
func (m *CustomPocModel) FindByTemplateId(ctx context.Context, templateId string) (*CustomPoc, error) {
	var doc CustomPoc
	err := m.coll.FindOne(ctx, bson.M{"template_id": templateId}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (m *CustomPocModel) Count(ctx context.Context) (int64, error) {
	return m.coll.CountDocuments(ctx, bson.M{})
}

func (m *CustomPocModel) Update(ctx context.Context, id string, doc *CustomPoc) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update := bson.M{
		"name":        doc.Name,
		"template_id": doc.TemplateId,
		"severity":    doc.Severity,
		"tags":        doc.Tags,
		"author":      doc.Author,
		"description": doc.Description,
		"content":     doc.Content,
		"enabled":     doc.Enabled,
		"update_time": time.Now(),
	}
	_, err = m.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
	return err
}

func (m *CustomPocModel) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

// DeleteAll 删除所有自定义POC
func (m *CustomPocModel) DeleteAll(ctx context.Context) (int64, error) {
	result, err := m.coll.DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// FindByIds 根据ID列表获取自定义POC
func (m *CustomPocModel) FindByIds(ctx context.Context, ids []string) ([]CustomPoc, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	// 转换字符串ID为ObjectID
	oids := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		oids = append(oids, oid)
	}

	if len(oids) == 0 {
		return nil, nil
	}

	cursor, err := m.coll.Find(ctx, bson.M{"_id": bson.M{"$in": oids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []CustomPoc
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// NucleiTemplate Nuclei默认模板（从模板目录同步）
type NucleiTemplate struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TemplateId  string             `bson:"template_id" json:"templateId"`   // 模板ID
	Name        string             `bson:"name" json:"name"`                // 模板名称
	Author      string             `bson:"author" json:"author"`            // 作者
	Severity    string             `bson:"severity" json:"severity"`        // 严重级别
	Description string             `bson:"description" json:"description"`  // 描述
	Tags        []string           `bson:"tags" json:"tags"`                // 标签
	Category    string             `bson:"category" json:"category"`        // 分类(目录名)
	FilePath    string             `bson:"file_path" json:"filePath"`       // 相对文件路径
	Content     string             `bson:"content" json:"content"`          // YAML内容
	Enabled     bool               `bson:"enabled" json:"enabled"`          // 是否启用
	SyncTime    time.Time          `bson:"sync_time" json:"syncTime"`       // 同步时间

	// 漏洞知识库字段
	CvssScore   float64  `bson:"cvss_score,omitempty" json:"cvssScore,omitempty"`     // CVSS评分
	CvssMetrics string   `bson:"cvss_metrics,omitempty" json:"cvssMetrics,omitempty"` // CVSS向量
	CveIds      []string `bson:"cve_ids,omitempty" json:"cveIds,omitempty"`           // CVE编号列表
	CweIds      []string `bson:"cwe_ids,omitempty" json:"cweIds,omitempty"`           // CWE编号列表
	References  []string `bson:"references,omitempty" json:"references,omitempty"`   // 参考链接
	Remediation string   `bson:"remediation,omitempty" json:"remediation,omitempty"` // 修复建议
}

// NucleiTemplateModel Nuclei模板模型
type NucleiTemplateModel struct {
	coll *mongo.Collection
}

func NewNucleiTemplateModel(db *mongo.Database) *NucleiTemplateModel {
	coll := db.Collection("nuclei_template")
	// 创建索引
	coll.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "template_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "category", Value: 1}}},
		{Keys: bson.D{{Key: "severity", Value: 1}}},
		{Keys: bson.D{{Key: "tags", Value: 1}}},
		{Keys: bson.D{{Key: "name", Value: "text"}, {Key: "template_id", Value: "text"}, {Key: "description", Value: "text"}}},
		// 支持CVSS和CVE查询的索引
		{Keys: bson.D{{Key: "cvss_score", Value: -1}}},
		{Keys: bson.D{{Key: "cve_ids", Value: 1}}},
	})
	return &NucleiTemplateModel{coll: coll}
}

func (m *NucleiTemplateModel) Upsert(ctx context.Context, doc *NucleiTemplate) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	doc.SyncTime = time.Now()
	
	filter := bson.M{"template_id": doc.TemplateId}
	update := bson.M{"$set": doc}
	opts := options.Update().SetUpsert(true)
	_, err := m.coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func (m *NucleiTemplateModel) BulkUpsert(ctx context.Context, docs []*NucleiTemplate) error {
	if len(docs) == 0 {
		return nil
	}
	
	var models []mongo.WriteModel
	now := time.Now()
	for _, doc := range docs {
		if doc.Id.IsZero() {
			doc.Id = primitive.NewObjectID()
		}
		doc.SyncTime = now
		
		filter := bson.M{"template_id": doc.TemplateId}
		update := bson.M{"$set": doc}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
	}
	
	opts := options.BulkWrite().SetOrdered(false)
	_, err := m.coll.BulkWrite(ctx, models, opts)
	return err
}

func (m *NucleiTemplateModel) Find(ctx context.Context, filter bson.M, page, pageSize int) ([]NucleiTemplate, error) {
	opts := options.Find()
	if page > 0 && pageSize > 0 {
		opts.SetSkip(int64((page - 1) * pageSize))
		opts.SetLimit(int64(pageSize))
	}
	opts.SetSort(bson.D{{Key: "severity", Value: 1}, {Key: "name", Value: 1}})
	// 排除content字段，提高查询性能
	opts.SetProjection(bson.M{"content": 0})

	cursor, err := m.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []NucleiTemplate
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *NucleiTemplateModel) Count(ctx context.Context, filter bson.M) (int64, error) {
	return m.coll.CountDocuments(ctx, filter)
}

// FindByTemplateId 根据模板ID获取完整模板（包含content）
func (m *NucleiTemplateModel) FindByTemplateId(ctx context.Context, templateId string) (*NucleiTemplate, error) {
	var doc NucleiTemplate
	err := m.coll.FindOne(ctx, bson.M{"template_id": templateId}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (m *NucleiTemplateModel) GetCategories(ctx context.Context) ([]string, error) {
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

func (m *NucleiTemplateModel) GetTags(ctx context.Context, limit int) ([]string, error) {
	pipeline := []bson.M{
		{"$unwind": "$tags"},
		{"$group": bson.M{"_id": "$tags", "count": bson.M{"$sum": 1}}},
		{"$sort": bson.M{"count": -1}},
	}
	
	// 只有当limit > 0时才添加限制
	if limit > 0 {
		pipeline = append(pipeline, bson.M{"$limit": limit})
	}
	
	cursor, err := m.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var results []struct {
		Id    string `bson:"_id"`
		Count int    `bson:"count"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	
	tags := make([]string, 0, len(results))
	for _, r := range results {
		if r.Id != "" { // 过滤空标签
			tags = append(tags, r.Id)
		}
	}
	return tags, nil
}

func (m *NucleiTemplateModel) GetStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	// 使用聚合管道一次性统计所有严重级别
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id":   "$severity",
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := m.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return stats, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Id    string `bson:"_id"`
		Count int    `bson:"count"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return stats, err
	}

	total := 0
	for _, r := range results {
		if r.Id != "" {
			stats[r.Id] = r.Count
			total += r.Count
		}
	}
	stats["total"] = total

	return stats, nil
}

func (m *NucleiTemplateModel) DeleteAll(ctx context.Context) error {
	_, err := m.coll.DeleteMany(ctx, bson.M{})
	return err
}

// FindEnabled 获取启用的模板
func (m *NucleiTemplateModel) FindEnabled(ctx context.Context) ([]NucleiTemplate, error) {
	cursor, err := m.coll.Find(ctx, bson.M{"enabled": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []NucleiTemplate
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// FindEnabledByFilter 根据条件获取启用的模板
func (m *NucleiTemplateModel) FindEnabledByFilter(ctx context.Context, filter bson.M) ([]NucleiTemplate, error) {
	filter["enabled"] = true
	cursor, err := m.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []NucleiTemplate
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// FindBySeverity 根据严重级别获取启用的模板
func (m *NucleiTemplateModel) FindBySeverity(ctx context.Context, severities []string) ([]NucleiTemplate, error) {
	filter := bson.M{
		"enabled":  true,
		"severity": bson.M{"$in": severities},
	}
	cursor, err := m.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []NucleiTemplate
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// FindByTags 根据标签获取启用的模板
func (m *NucleiTemplateModel) FindByTags(ctx context.Context, tags []string) ([]NucleiTemplate, error) {
	filter := bson.M{
		"enabled": true,
		"tags":    bson.M{"$in": tags},
	}
	cursor, err := m.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []NucleiTemplate
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// UpdateEnabled 更新模板启用状态
func (m *NucleiTemplateModel) UpdateEnabled(ctx context.Context, templateId string, enabled bool) error {
	_, err := m.coll.UpdateOne(ctx, bson.M{"template_id": templateId}, bson.M{"$set": bson.M{"enabled": enabled}})
	return err
}

// BatchUpdateEnabled 批量更新模板启用状态
func (m *NucleiTemplateModel) BatchUpdateEnabled(ctx context.Context, templateIds []string, enabled bool) error {
	_, err := m.coll.UpdateMany(ctx, bson.M{"template_id": bson.M{"$in": templateIds}}, bson.M{"$set": bson.M{"enabled": enabled}})
	return err
}

// FindByIds 根据ID列表获取模板（包含content）
func (m *NucleiTemplateModel) FindByIds(ctx context.Context, ids []string) ([]NucleiTemplate, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	// 转换字符串ID为ObjectID
	oids := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		oids = append(oids, oid)
	}

	if len(oids) == 0 {
		return nil, nil
	}

	cursor, err := m.coll.Find(ctx, bson.M{"_id": bson.M{"$in": oids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []NucleiTemplate
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}
