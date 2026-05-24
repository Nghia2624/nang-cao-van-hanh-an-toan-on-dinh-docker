package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureIndexes creates TTL and compound indexes for collections.
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	indexes := []struct {
		Collection string
		Models     []mongo.IndexModel
	}{
		{
			Collection: "logs",
			Models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "timestamp", Value: -1}}},
				{Keys: bson.D{{Key: "container_id", Value: 1}, {Key: "timestamp", Value: -1}}},
				{Keys: bson.D{{Key: "fingerprint", Value: 1}, {Key: "timestamp", Value: -1}}},
				{
					Keys:    bson.D{{Key: "expireAt", Value: 1}},
					Options: options.Index().SetExpireAfterSeconds(0),
				},
			},
		},
		{
			Collection: "ai_analysis",
			Models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "fingerprint", Value: 1}, {Key: "created_at", Value: -1}}},
				{
					Keys:    bson.D{{Key: "expireAt", Value: 1}},
					Options: options.Index().SetExpireAfterSeconds(0),
				},
			},
		},
		{
			Collection: "alerts",
			Models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}}},
				{Keys: bson.D{{Key: "fingerprint", Value: 1}}},
			},
		},
		{
			Collection: "prediction_feedback",
			Models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "container_id", Value: 1}, {Key: "created_at", Value: -1}}},
				{Keys: bson.D{{Key: "accurate", Value: 1}, {Key: "created_at", Value: -1}}},
				{
					Keys:    bson.D{{Key: "expireAt", Value: 1}},
					Options: options.Index().SetExpireAfterSeconds(0),
				},
			},
		},
		{
			Collection: "ai_chat_sessions",
			Models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "updated_at", Value: -1}}},
				{Keys: bson.D{{Key: "created_at", Value: -1}}},
				{
					Keys:    bson.D{{Key: "expireAt", Value: 1}},
					Options: options.Index().SetExpireAfterSeconds(0),
				},
			},
		},
	}

	for _, item := range indexes {
		collection := db.Collection(item.Collection)
		if _, err := collection.Indexes().CreateMany(ctx, item.Models); err != nil {
			return err
		}
	}
	return nil
}

// WithTTL sets a future expiration date on a document (to be used by caller).
func WithTTL(hours int) time.Time {
	return time.Now().Add(time.Duration(hours) * time.Hour)
}
