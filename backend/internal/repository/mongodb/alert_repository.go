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

type AlertRepo struct {
	col *mongo.Collection
}

func NewAlertRepo(client *Client) *AlertRepo {
	return &AlertRepo{col: client.Collection("alerts")}
}

func (r *AlertRepo) Save(ctx context.Context, alert *domain.Alert) error {
	doc := bson.M{
		"source":       alert.Source,
		"severity":     alert.Severity,
		"title":        alert.Title,
		"description":  alert.Description,
		"status":       alert.Status,
		"container_id": alert.ContainerID,
		"fingerprint":  alert.Fingerprint,
		"created_at":   alert.CreatedAt,
		"updated_at":   alert.UpdatedAt,
	}
	_, err := r.col.InsertOne(ctx, doc)
	return err
}

func (r *AlertRepo) UpdateStatus(ctx context.Context, id string, status domain.AlertStatus) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.col.UpdateByID(ctx, objID, bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	})
	return err
}

func (r *AlertRepo) ExistsByFingerprintSince(ctx context.Context, fingerprint string, since time.Time) (bool, error) {
	if fingerprint == "" {
		return false, nil
	}
	filter := bson.M{
		"fingerprint": fingerprint,
		"created_at":  bson.M{"$gte": since},
		"status":      bson.M{"$in": []domain.AlertStatus{domain.AlertStatusNew, domain.AlertStatusAcknowledged}},
	}
	opts := options.FindOne().SetProjection(bson.M{"_id": 1})
	var out bson.M
	err := r.col.FindOne(ctx, filter, opts).Decode(&out)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *AlertRepo) List(ctx context.Context, status domain.AlertStatus, limit int) ([]*domain.Alert, error) {
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(int64(limit))
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var res []*domain.Alert
	for cur.Next(ctx) {
		var row struct {
			ID          primitive.ObjectID `bson:"_id"`
			Source      string             `bson:"source"`
			Severity    string             `bson:"severity"`
			Title       string             `bson:"title"`
			Description string             `bson:"description"`
			Status      domain.AlertStatus `bson:"status"`
			ContainerID string             `bson:"container_id"`
			Fingerprint string             `bson:"fingerprint"`
			CreatedAt   time.Time          `bson:"created_at"`
			UpdatedAt   time.Time          `bson:"updated_at"`
		}
		if err := cur.Decode(&row); err != nil {
			continue // Skip invalid entries
		}
		res = append(res, &domain.Alert{
			ID:          row.ID.Hex(), // Convert ObjectID to string
			Source:      row.Source,
			Severity:    row.Severity,
			Title:       row.Title,
			Description: row.Description,
			Status:      row.Status,
			ContainerID: row.ContainerID,
			Fingerprint: row.Fingerprint,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}
	if err := cur.Err(); err != nil {
		return res, err
	}
	if res == nil {
		return []*domain.Alert{}, nil
	}
	return res, nil
}
