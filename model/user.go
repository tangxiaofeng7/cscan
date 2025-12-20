package model

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	RoleAdmin    = "admin"
	RoleSuperAdmin = "superadmin"
	RoleGuest    = "guest"
	
	StatusEnable  = "enable"
	StatusDisable = "disable"
)

type User struct {
	Id              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username        string             `bson:"username" json:"username"`
	Password        string             `bson:"password" json:"-"`
	Role            string             `bson:"role" json:"role"`
	Status          string             `bson:"status" json:"status"`
	WorkspaceIds    []string           `bson:"workspace_ids" json:"workspaceIds"`
	LastLoginTime   *time.Time         `bson:"last_login_time" json:"lastLoginTime"`
	CreateTime      time.Time          `bson:"create_time" json:"createTime"`
	UpdateTime      time.Time          `bson:"update_time" json:"updateTime"`
}

type UserModel struct {
	coll *mongo.Collection
}

func NewUserModel(db *mongo.Database) *UserModel {
	return &UserModel{
		coll: db.Collection("user"),
	}
}

func (m *UserModel) Insert(ctx context.Context, doc *User) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	doc.Password = HashPassword(doc.Password)
	_, err := m.coll.InsertOne(ctx, doc)
	return err
}

func (m *UserModel) FindByUsername(ctx context.Context, username string) (*User, error) {
	var doc User
	err := m.coll.FindOne(ctx, bson.M{"username": username}).Decode(&doc)
	return &doc, err
}

func (m *UserModel) FindById(ctx context.Context, id string) (*User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var doc User
	err = m.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc)
	return &doc, err
}

func (m *UserModel) Find(ctx context.Context, filter bson.M, page, pageSize int) ([]User, error) {
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

	var docs []User
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (m *UserModel) Count(ctx context.Context, filter bson.M) (int64, error) {
	return m.coll.CountDocuments(ctx, filter)
}

func (m *UserModel) Update(ctx context.Context, id string, update bson.M) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update["update_time"] = time.Now()
	_, err = m.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
	return err
}

func (m *UserModel) UpdateLoginTime(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = m.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"last_login_time": now}})
	return err
}

func (m *UserModel) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

func (m *UserModel) VerifyPassword(ctx context.Context, username, password string) (*User, bool) {
	user, err := m.FindByUsername(ctx, username)
	if err != nil {
		return nil, false
	}
	if user.Password != HashPassword(password) {
		return nil, false
	}
	if user.Status != StatusEnable {
		return nil, false
	}
	return user, true
}

func HashPassword(password string) string {
	hash := md5.Sum([]byte(password))
	return hex.EncodeToString(hash[:])
}
