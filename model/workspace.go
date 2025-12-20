package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Workspace struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Status      string             `bson:"status" json:"status"`
	CreateTime  time.Time          `bson:"create_time" json:"createTime"`
	UpdateTime  time.Time          `bson:"update_time" json:"updateTime"`
}

type WorkspaceModel struct {
	coll *mongo.Collection
}

func NewWorkspaceModel(db *mongo.Database) *WorkspaceModel {
	return &WorkspaceModel{
		coll: db.Collection("workspace"),
	}
}

func (m *WorkspaceModel) Insert(ctx context.Context, doc *Workspace) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	doc.Status = StatusEnable
	_, err := m.coll.InsertOne(ctx, doc)
	return err
}

func (m *WorkspaceModel) FindById(ctx context.Context, id string) (*Workspace, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var doc Workspace
	err = m.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	return &doc, err
}

func (m *WorkspaceModel) Find(ctx context.Context, filter bson.M, page, pageSize int) ([]Workspace, error) {
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

	var docs []Workspace
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *WorkspaceModel) FindByIds(ctx context.Context, ids []string) ([]Workspace, error) {
	var oids []primitive.ObjectID
	for _, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		oids = append(oids, oid)
	}
	
	cursor, err := m.coll.Find(ctx, bson.M{"_id": bson.M{"$in": oids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []Workspace
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *WorkspaceModel) Count(ctx context.Context, filter bson.M) (int64, error) {
	return m.coll.CountDocuments(ctx, filter)
}

func (m *WorkspaceModel) Update(ctx context.Context, id string, update bson.M) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update["update_time"] = time.Now()
	_, err = m.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
	return err
}

func (m *WorkspaceModel) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}
