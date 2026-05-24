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

type AIRepo struct {
	col *mongo.Collection
}

func NewAIRepo(client *Client) *AIRepo {
	return &AIRepo{col: client.Collection("ai_analysis")}
}

func (r *AIRepo) Save(ctx context.Context, analysis *domain.AIAnalysis) error {
	doc := bson.M{
		"fingerprint":         analysis.Fingerprint,
		"root_cause":          analysis.RootCause,
		"severity":            analysis.Severity,
		"severity_label":      analysis.SeverityLabel,
		"impact_analysis":     analysis.ImpactAnalysis,
		"affected_components": analysis.AffectedComponents,
		"recommended_actions": analysis.RecommendedActions,
		"prevention_measures": analysis.PreventionMeasures,
		"related_issues":      analysis.RelatedIssues,
		"confidence_score":    analysis.ConfidenceScore,
		"created_at":          analysis.CreatedAt,
		"expireAt":            WithTTL(24 * 30),
	}
	result, err := r.col.InsertOne(ctx, doc)
	if err != nil {
	return err
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		analysis.ID = oid.Hex()
	}
	return nil
}

func (r *AIRepo) FindByFingerprint(ctx context.Context, fingerprint string, since time.Time) (*domain.AIAnalysis, error) {
	filter := bson.M{
		"fingerprint": fingerprint,
		"created_at":  bson.M{"$gte": since},
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	var row struct {
		ID                 primitive.ObjectID `bson:"_id"`
		Fingerprint        string             `bson:"fingerprint"`
		RootCause          string             `bson:"root_cause"`
		Severity           int                `bson:"severity"`
		SeverityLabel      string             `bson:"severity_label"`
		ImpactAnalysis     string             `bson:"impact_analysis"`
		AffectedComponents []string           `bson:"affected_components"`
		RecommendedActions []domain.Action   `bson:"recommended_actions"`
		PreventionMeasures []string           `bson:"prevention_measures"`
		RelatedIssues      []string           `bson:"related_issues"`
		ConfidenceScore    float64            `bson:"confidence_score"`
		CreatedAt          time.Time          `bson:"created_at"`
	}
	if err := r.col.FindOne(ctx, filter, opts).Decode(&row); err != nil {
		return nil, err
	}
	return &domain.AIAnalysis{
		ID:                 row.ID.Hex(),
		Fingerprint:        row.Fingerprint,
		RootCause:          row.RootCause,
		Severity:           row.Severity,
		SeverityLabel:      row.SeverityLabel,
		ImpactAnalysis:     row.ImpactAnalysis,
		AffectedComponents: row.AffectedComponents,
		RecommendedActions: row.RecommendedActions,
		PreventionMeasures: row.PreventionMeasures,
		RelatedIssues:      row.RelatedIssues,
		ConfidenceScore:    row.ConfidenceScore,
		CreatedAt:          row.CreatedAt,
	}, nil
}

func (r *AIRepo) List(ctx context.Context, limit int) ([]*domain.AIAnalysis, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(int64(limit))
	cur, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return []*domain.AIAnalysis{}, err
	}
	defer cur.Close(ctx)

	var res []*domain.AIAnalysis
	for cur.Next(ctx) {
		var row struct {
			ID                 primitive.ObjectID `bson:"_id"`
			Fingerprint        string             `bson:"fingerprint"`
			RootCause          string             `bson:"root_cause"`
			Severity           int                `bson:"severity"`
			SeverityLabel      string             `bson:"severity_label"`
			ImpactAnalysis     string             `bson:"impact_analysis"`
			AffectedComponents []string           `bson:"affected_components"`
			RecommendedActions []domain.Action   `bson:"recommended_actions"`
			PreventionMeasures []string           `bson:"prevention_measures"`
			RelatedIssues      []string           `bson:"related_issues"`
			ConfidenceScore    float64            `bson:"confidence_score"`
			CreatedAt          time.Time          `bson:"created_at"`
		}
		if err := cur.Decode(&row); err != nil {
			continue // Skip invalid entries instead of failing
		}
		res = append(res, &domain.AIAnalysis{
			ID:                 row.ID.Hex(),
			Fingerprint:        row.Fingerprint,
			RootCause:          row.RootCause,
			Severity:           row.Severity,
			SeverityLabel:      row.SeverityLabel,
			ImpactAnalysis:     row.ImpactAnalysis,
			AffectedComponents: row.AffectedComponents,
			RecommendedActions: row.RecommendedActions,
			PreventionMeasures: row.PreventionMeasures,
			RelatedIssues:      row.RelatedIssues,
			ConfidenceScore:    row.ConfidenceScore,
			CreatedAt:          row.CreatedAt,
		})
	}
	if err := cur.Err(); err != nil {
		return res, err
	}
	// Always return at least empty array, never nil
	if res == nil {
		return []*domain.AIAnalysis{}, nil
	}
	return res, nil
}
