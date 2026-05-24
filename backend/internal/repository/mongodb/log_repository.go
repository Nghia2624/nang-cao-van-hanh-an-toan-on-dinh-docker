package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/nghia/dockerai/backend/internal/domain"
)

type LogRepo struct {
	col *mongo.Collection
}

func NewLogRepo(client *Client) *LogRepo {
	return &LogRepo{col: client.Collection("logs")}
}

func (r *LogRepo) Insert(ctx context.Context, entry *domain.LogEntry) error {
	doc := bson.M{
		"container_id":  entry.ContainerID,
		"container":     entry.Container,
		"image":         entry.Image,
		"level":         entry.Level,
		"message":       entry.Message,
		"timestamp":     entry.Timestamp,
		"stream":        entry.Stream,
		"fingerprint":   entry.Fingerprint,
		"labels":        entry.Labels,
		"request_id":    entry.RequestID,
		"deployment_id": entry.DeploymentID,
		"expireAt":      WithTTL(24 * 30), // 30 days default
	}
	_, err := r.col.InsertOne(ctx, doc)
	return err
}

func (r *LogRepo) InsertBatch(ctx context.Context, entries []*domain.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	batch := make([]interface{}, 0, len(entries))
	for _, e := range entries {
		batch = append(batch, bson.M{
			"container_id":  e.ContainerID,
			"container":     e.Container,
			"image":         e.Image,
			"level":         e.Level,
			"message":       e.Message,
			"timestamp":     e.Timestamp,
			"stream":        e.Stream,
			"fingerprint":   e.Fingerprint,
			"labels":        e.Labels,
			"request_id":    e.RequestID,
			"deployment_id": e.DeploymentID,
			"expireAt":      WithTTL(24 * 30),
		})
	}
	_, err := r.col.InsertMany(ctx, batch)
	return err
}

func (r *LogRepo) ExistsByFingerprint(ctx context.Context, fingerprint string, since time.Time) (bool, error) {
	if fingerprint == "" {
		return false, nil
	}
	filter := bson.M{"fingerprint": fingerprint, "timestamp": bson.M{"$gte": since}}
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

func (r *LogRepo) Query(ctx context.Context, containerID string, from, to time.Time, level domain.LogLevel, limit int) ([]*domain.LogEntry, error) {
	filter := bson.M{
		"timestamp": bson.M{"$gte": from, "$lte": to},
	}
	if containerID != "" {
		filter["container_id"] = containerID
	}
	if level != "" {
		filter["level"] = level
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit))

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var res []*domain.LogEntry
	for cur.Next(ctx) {
		var row struct {
			ContainerID  string            `bson:"container_id"`
			Container    string            `bson:"container"`
			Image        string            `bson:"image"`
			Level        domain.LogLevel   `bson:"level"`
			Message      string            `bson:"message"`
			Timestamp    time.Time         `bson:"timestamp"`
			Stream       string            `bson:"stream"`
			Fingerprint  string            `bson:"fingerprint"`
			Labels       map[string]string `bson:"labels"`
			RequestID    string            `bson:"request_id"`
			DeploymentID string            `bson:"deployment_id"`
		}
		if err := cur.Decode(&row); err != nil {
			continue // Skip invalid entries
		}
		res = append(res, &domain.LogEntry{
			ContainerID:  row.ContainerID,
			Container:    row.Container,
			Image:        row.Image,
			Level:        row.Level,
			Message:      row.Message,
			Timestamp:    row.Timestamp,
			Stream:       row.Stream,
			Fingerprint:  row.Fingerprint,
			Labels:       row.Labels,
			RequestID:    row.RequestID,
			DeploymentID: row.DeploymentID,
		})
	}
	if err := cur.Err(); err != nil {
		return res, err
	}
	if res == nil {
		return []*domain.LogEntry{}, nil
	}
	return res, nil
}
