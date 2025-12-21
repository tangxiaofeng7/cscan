package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 任务状态常量
const (
	TaskStatusCreated  = "CREATED"  // 已创建，等待手动启动
	TaskStatusPending  = "PENDING"  // 等待执行（已入队）
	TaskStatusStarted  = "STARTED"  // 执行中
	TaskStatusPaused   = "PAUSED"   // 已暂停
	TaskStatusSuccess  = "SUCCESS"  // 执行成功
	TaskStatusFailure  = "FAILURE"  // 执行失败
	TaskStatusRevoked  = "REVOKED"  // 已取消
	TaskStatusStopped  = "STOPPED"  // 已停止
)

type MainTask struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TaskId      string             `bson:"task_id" json:"taskId"`
	Name        string             `bson:"name" json:"name"`
	Target      string             `bson:"target" json:"target"`
	ProfileId   string             `bson:"profile_id" json:"profileId"`
	ProfileName string             `bson:"profile_name" json:"profileName"`
	Status      string             `bson:"status" json:"status"`
	Progress    int                `bson:"progress" json:"progress"`
	Result      string             `bson:"result" json:"result"`
	IsCron      bool               `bson:"is_cron" json:"isCron"`
	CronRule    string             `bson:"cron_rule" json:"cronRule"`
	CronStatus  string             `bson:"cron_status" json:"cronStatus"`
	NotifyId    string             `bson:"notify_id" json:"notifyId"`
	CreateTime  time.Time          `bson:"create_time" json:"createTime"`
	UpdateTime  time.Time          `bson:"update_time" json:"updateTime"`
	StartTime   *time.Time         `bson:"start_time" json:"startTime"`
	EndTime     *time.Time         `bson:"end_time" json:"endTime"`
	// 任务进度保存（用于暂停/继续）
	TaskState   string             `bson:"task_state" json:"taskState"`     // 任务执行状态JSON（保存已完成的阶段和数据）
	Config      string             `bson:"config" json:"config"`            // 任务配置JSON
}

type ExecutorTask struct {
	Id         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TaskId     string             `bson:"task_id" json:"taskId"`
	MainTaskId string             `bson:"main_task_id" json:"mainTaskId"`
	TaskName   string             `bson:"task_name" json:"taskName"`
	Config     string             `bson:"config" json:"config"`
	Status     string             `bson:"status" json:"status"`
	Worker     string             `bson:"worker" json:"worker"`
	Result     string             `bson:"result" json:"result"`
	CreateTime time.Time          `bson:"create_time" json:"createTime"`
	StartTime  *time.Time         `bson:"start_time" json:"startTime"`
	EndTime    *time.Time         `bson:"end_time" json:"endTime"`
}

type TaskProfile struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Config      string             `bson:"config" json:"config"`
	SortNumber  int                `bson:"sort_number" json:"sortNumber"`
	CreateTime  time.Time          `bson:"create_time" json:"createTime"`
	UpdateTime  time.Time          `bson:"update_time" json:"updateTime"`
}

type MainTaskModel struct {
	coll *mongo.Collection
}

func NewMainTaskModel(db *mongo.Database, workspaceId string) *MainTaskModel {
	coll := db.Collection(workspaceId + "_maintask")

	// 创建索引
	ctx := context.Background()
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "task_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "create_time", Value: -1}}},
	}
	coll.Indexes().CreateMany(ctx, indexes)

	return &MainTaskModel{
		coll: coll,
	}
}

func (m *MainTaskModel) Insert(ctx context.Context, doc *MainTask) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	doc.Status = TaskStatusCreated
	_, err := m.coll.InsertOne(ctx, doc)
	return err
}

func (m *MainTaskModel) FindById(ctx context.Context, id string) (*MainTask, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var doc MainTask
	err = m.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	return &doc, err
}

func (m *MainTaskModel) FindByTaskId(ctx context.Context, taskId string) (*MainTask, error) {
	var doc MainTask
	err := m.coll.FindOne(ctx, bson.M{"task_id": taskId}).Decode(&doc)
	return &doc, err
}

func (m *MainTaskModel) Find(ctx context.Context, filter bson.M, page, pageSize int) ([]MainTask, error) {
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

	var docs []MainTask
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *MainTaskModel) Count(ctx context.Context, filter bson.M) (int64, error) {
	return m.coll.CountDocuments(ctx, filter)
}

func (m *MainTaskModel) Update(ctx context.Context, id string, update bson.M) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update["update_time"] = time.Now()
	_, err = m.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
	return err
}

func (m *MainTaskModel) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

func (m *MainTaskModel) BatchDelete(ctx context.Context, ids []string) (int64, error) {
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

// UpdateByTaskId 根据 taskId 更新任务
func (m *MainTaskModel) UpdateByTaskId(ctx context.Context, taskId string, update bson.M) error {
	update["update_time"] = time.Now()
	_, err := m.coll.UpdateOne(ctx, bson.M{"task_id": taskId}, bson.M{"$set": update})
	return err
}

// ExecutorTaskModel
type ExecutorTaskModel struct {
	coll *mongo.Collection
}

func NewExecutorTaskModel(db *mongo.Database, workspaceId string) *ExecutorTaskModel {
	return &ExecutorTaskModel{
		coll: db.Collection(workspaceId + "_executor_task"),
	}
}

func (m *ExecutorTaskModel) Insert(ctx context.Context, doc *ExecutorTask) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	doc.CreateTime = time.Now()
	doc.Status = TaskStatusPending
	_, err := m.coll.InsertOne(ctx, doc)
	return err
}

func (m *ExecutorTaskModel) FindByMainTaskId(ctx context.Context, mainTaskId string, page, pageSize int) ([]ExecutorTask, error) {
	opts := options.Find()
	if page > 0 && pageSize > 0 {
		opts.SetSkip(int64((page - 1) * pageSize))
		opts.SetLimit(int64(pageSize))
	}
	opts.SetSort(bson.D{{Key: "create_time", Value: -1}})

	cursor, err := m.coll.Find(ctx, bson.M{"main_task_id": mainTaskId}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []ExecutorTask
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// TaskProfileModel
type TaskProfileModel struct {
	coll *mongo.Collection
}

func NewTaskProfileModel(db *mongo.Database) *TaskProfileModel {
	return &TaskProfileModel{
		coll: db.Collection("task_profile"),
	}
}

func (m *TaskProfileModel) FindAll(ctx context.Context) ([]TaskProfile, error) {
	opts := options.Find().SetSort(bson.D{{Key: "sort_number", Value: 1}})
	cursor, err := m.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []TaskProfile
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *TaskProfileModel) FindById(ctx context.Context, id string) (*TaskProfile, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var doc TaskProfile
	err = m.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	return &doc, err
}

func (m *TaskProfileModel) Insert(ctx context.Context, doc *TaskProfile) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	_, err := m.coll.InsertOne(ctx, doc)
	return err
}

func (m *TaskProfileModel) Update(ctx context.Context, id string, doc *TaskProfile) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update := bson.M{
		"name":        doc.Name,
		"description": doc.Description,
		"config":      doc.Config,
		"update_time": time.Now(),
	}
	_, err = m.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
	return err
}

func (m *TaskProfileModel) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}
