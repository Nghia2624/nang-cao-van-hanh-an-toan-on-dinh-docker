package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/nghia/dockerai/backend/internal/domain"
)

type AIChatRepo struct {
	col *mongo.Collection
}

func NewAIChatRepo(client *Client) *AIChatRepo {
	return &AIChatRepo{col: client.Collection("ai_chat_sessions")}
}

func (r *AIChatRepo) CreateSession(ctx context.Context, title string) (*domain.AIChatSession, error) {
	now := time.Now().UTC()
	s := &domain.AIChatSession{
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  []domain.AIChatMessage{},
		ExpireAt:  WithTTL(24 * 30),
	}
	doc := bson.M{
		"title":      s.Title,
		"created_at": s.CreatedAt,
		"updated_at": s.UpdatedAt,
		"messages":   []interface{}{},
		"expireAt":   s.ExpireAt,
	}
	res, err := r.col.InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		s.ID = oid.Hex()
	}
	return s, nil
}

func (r *AIChatRepo) GetSession(ctx context.Context, id string) (*domain.AIChatSession, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": oid}
	var row struct {
		ID        primitive.ObjectID     `bson:"_id"`
		Title     string                 `bson:"title"`
		CreatedAt time.Time              `bson:"created_at"`
		UpdatedAt time.Time              `bson:"updated_at"`
		Messages  []domain.AIChatMessage `bson:"messages"`
		ExpireAt  time.Time              `bson:"expireAt"`
	}
	if err := r.col.FindOne(ctx, filter).Decode(&row); err != nil {
		return nil, err
	}
	if row.Messages == nil {
		row.Messages = []domain.AIChatMessage{}
	}
	return &domain.AIChatSession{
		ID:        row.ID.Hex(),
		Title:     row.Title,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		Messages:  row.Messages,
		ExpireAt:  row.ExpireAt,
	}, nil
}

func (r *AIChatRepo) AppendMessage(ctx context.Context, sessionID string, msg domain.AIChatMessage) error {
	oid, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return err
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now().UTC()
	}
	update := bson.M{
		"$push": bson.M{"messages": msg},
		"$set":  bson.M{"updated_at": time.Now().UTC(), "expireAt": WithTTL(24 * 30)},
	}
	_, err = r.col.UpdateByID(ctx, oid, update)
	return err
}

func (r *AIChatRepo) SetTitle(ctx context.Context, sessionID string, title string) error {
	oid, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return err
	}
	update := bson.M{
		"$set": bson.M{
			"title":      title,
			"updated_at": time.Now().UTC(),
		},
	}
	_, err = r.col.UpdateByID(ctx, oid, update)
	return err
}

func (r *AIChatRepo) ListSessions(ctx context.Context, limit int) ([]*domain.AIChatSession, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}}).SetLimit(int64(limit))
	cur, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return []*domain.AIChatSession{}, err
	}
	defer cur.Close(ctx)

	res := []*domain.AIChatSession{}
	for cur.Next(ctx) {
		var row struct {
			ID        primitive.ObjectID `bson:"_id"`
			Title     string             `bson:"title"`
			CreatedAt time.Time          `bson:"created_at"`
			UpdatedAt time.Time          `bson:"updated_at"`
			ExpireAt  time.Time          `bson:"expireAt"`
		}
		if err := cur.Decode(&row); err != nil {
			continue
		}
		res = append(res, &domain.AIChatSession{
			ID:        row.ID.Hex(),
			Title:     row.Title,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			Messages:  []domain.AIChatMessage{},
			ExpireAt:  row.ExpireAt,
		})
	}
	if res == nil {
		return []*domain.AIChatSession{}, nil
	}
	return res, nil
}
