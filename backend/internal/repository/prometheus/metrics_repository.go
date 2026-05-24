package prometheus

import (
	"context"
	"fmt"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/nghia/dockerai/backend/internal/domain"
)

type MetricsRepo struct {
	api v1.API
}

func New(url string) (*MetricsRepo, error) {
	client, err := promapi.NewClient(promapi.Config{Address: url})
	if err != nil {
		return nil, err
	}
	return &MetricsRepo{api: v1.NewAPI(client)}, nil
}

func (r *MetricsRepo) QueryRange(ctx context.Context, promql string, from, to time.Time, step time.Duration) ([]domain.MetricPoint, error) {
	result, _, err := r.api.QueryRange(ctx, promql, v1.Range{Start: from, End: to, Step: step})
	if err != nil {
		return nil, err
	}
	matrix, ok := result.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unexpected result type %T", result)
	}
	var points []domain.MetricPoint
	for _, stream := range matrix {
		for _, v := range stream.Values {
			points = append(points, domain.MetricPoint{
				Timestamp: v.Timestamp.Time(),
				Value:     float64(v.Value),
			})
		}
	}
	return points, nil
}

func (r *MetricsRepo) QueryInstant(ctx context.Context, promql string, ts time.Time) (float64, error) {
	result, _, err := r.api.Query(ctx, promql, ts)
	if err != nil {
		return 0, err
	}
	vector, ok := result.(model.Vector)
	if !ok || len(vector) == 0 {
		return 0, fmt.Errorf("no data")
	}
	return float64(vector[0].Value), nil
}
