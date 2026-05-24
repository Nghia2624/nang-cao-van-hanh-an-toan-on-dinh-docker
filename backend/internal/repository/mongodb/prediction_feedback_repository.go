package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/nghia/dockerai/backend/internal/domain"
)

type PredictionFeedbackRepo struct {
	col *mongo.Collection
}

func NewPredictionFeedbackRepo(client *Client) *PredictionFeedbackRepo {
	return &PredictionFeedbackRepo{col: client.Collection("prediction_feedback")}
}

func (r *PredictionFeedbackRepo) Save(ctx context.Context, fb *domain.PredictionFeedback) error {
	if fb == nil {
		return nil
	}
	if fb.CreatedAt.IsZero() {
		fb.CreatedAt = time.Now().UTC()
	}
	doc := bson.M{
		"container_id":    fb.ContainerID,
		"risk_level":      fb.RiskLevel,
		"risk_score":      fb.RiskScore,
		"eta_minutes":     fb.ETAMinutes,
		"accurate":        fb.Accurate,
		"note":            fb.Note,
		"created_at":      fb.CreatedAt,
		"prediction_from": fb.PredictionFrom,
		"prediction_to":   fb.PredictionTo,
		"expireAt":        WithTTL(24 * 30),
	}
	res, err := r.col.InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		fb.ID = oid.Hex()
	}
	return nil
}

func (r *PredictionFeedbackRepo) SaveChatFeedback(ctx context.Context, fb *domain.ChatPredictionFeedback) error {
	if fb == nil {
		return nil
	}
	if fb.CreatedAt.IsZero() {
		fb.CreatedAt = time.Now().UTC()
	}
	doc := bson.M{
		"prediction_id": fb.PredictionID,
		"container_id":  fb.ContainerID,
		"is_correct":    fb.IsCorrect,
		"comment":       fb.Comment,
		"created_at":    fb.CreatedAt,
		"expireAt":      WithTTL(24 * 30),
	}
	res, err := r.col.InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		fb.ID = oid.Hex()
	}
	return nil
}
