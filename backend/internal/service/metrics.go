package service

import (
	"context"
	"sort"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"github.com/google/uuid"
)

// MetricsService aggregates raw per-request telemetry into the request-count
// series and breakdowns rendered on the dashboard.
type MetricsService struct {
	requests repository.RequestMetricsRepository
	envs     repository.DashboardEnvironmentRepository
	// now is injected so tests can pin the clock; production uses time.Now.
	now func() time.Time
}

// NewMetricsService wires the request-metrics and environment repositories.
func NewMetricsService(requests repository.RequestMetricsRepository, envs repository.DashboardEnvironmentRepository) *MetricsService {
	return &MetricsService{requests: requests, envs: envs, now: time.Now}
}

// RequestMetricsBucket is a single point on the request-count time series.
type RequestMetricsBucket struct {
	Start   time.Time `json:"start"`
	Label   string    `json:"label"`
	Count   int       `json:"count"`
	Success int       `json:"success"`
	Failed  int       `json:"failed"`
}

// RequestMetricsEnvBreakdown counts requests routed through one environment.
type RequestMetricsEnvBreakdown struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Count  int    `json:"count"`
	Active bool   `json:"active"`
}

// RequestMetricsResult is the aggregated payload returned to the dashboard.
type RequestMetricsResult struct {
	Range        string                       `json:"range"`
	Bucket       string                       `json:"bucket"`
	Total        int                          `json:"total"`
	Success      int                          `json:"success"`
	Failed       int                          `json:"failed"`
	AvgLatencyMs int64                        `json:"avg_latency_ms"`
	Series       []RequestMetricsBucket       `json:"series"`
	Environments []RequestMetricsEnvBreakdown `json:"environments"`
}

// rangeSpec describes how a requested range maps to a bucketed series.
type rangeSpec struct {
	name     string
	bucket   string
	duration time.Duration
	step     time.Duration
	label    string
}

func specForRange(raw string) rangeSpec {
	switch raw {
	case "24h", "day", "1d":
		return rangeSpec{name: "24h", bucket: "hour", duration: 24 * time.Hour, step: time.Hour, label: "15:04"}
	case "30d", "month":
		return rangeSpec{name: "30d", bucket: "day", duration: 30 * 24 * time.Hour, step: 24 * time.Hour, label: "Jan 2"}
	case "7d", "week", "":
		return rangeSpec{name: "7d", bucket: "day", duration: 7 * 24 * time.Hour, step: 24 * time.Hour, label: "Jan 2"}
	default:
		return rangeSpec{name: "7d", bucket: "day", duration: 7 * 24 * time.Hour, step: 24 * time.Hour, label: "Jan 2"}
	}
}

// truncateTo snaps t down to the start of its bucket (hour or day) in UTC.
func truncateTo(t time.Time, step time.Duration) time.Time {
	t = t.UTC()
	if step >= 24*time.Hour {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}
	return t.Truncate(step)
}

// RequestMetrics returns the aggregated request metrics for a user, optionally
// scoped to a single environment, over the requested range ("24h"/"7d"/"30d").
func (s *MetricsService) RequestMetrics(ctx context.Context, userID uuid.UUID, envID *uuid.UUID, rangeRaw string) (*RequestMetricsResult, error) {
	spec := specForRange(rangeRaw)
	now := s.now().UTC()
	// Include the current, in-progress bucket, so align the window to bucket
	// boundaries and walk forward.
	end := truncateTo(now, spec.step).Add(spec.step)
	start := end.Add(-spec.duration)

	metrics, err := s.requests.ListByUserSince(ctx, userID, envID, start)
	if err != nil {
		return nil, err
	}

	// Pre-seed a contiguous set of empty buckets so the chart always has a
	// stable x-axis even for quiet periods.
	bucketCount := int(spec.duration / spec.step)
	buckets := make([]RequestMetricsBucket, bucketCount)
	index := make(map[time.Time]int, bucketCount)
	for i := 0; i < bucketCount; i++ {
		bStart := start.Add(time.Duration(i) * spec.step)
		buckets[i] = RequestMetricsBucket{Start: bStart, Label: bStart.Format(spec.label)}
		index[bStart] = i
	}

	envCounts := make(map[uuid.UUID]int)
	var total, success, failed int
	var latencySum int64
	for _, m := range metrics {
		total++
		if m.Success {
			success++
		} else {
			failed++
		}
		latencySum += m.LatencyMs

		bStart := truncateTo(m.CreatedAt, spec.step)
		if i, ok := index[bStart]; ok {
			buckets[i].Count++
			if m.Success {
				buckets[i].Success++
			} else {
				buckets[i].Failed++
			}
		}
		if m.EnvironmentID != nil {
			envCounts[*m.EnvironmentID]++
		}
	}

	result := &RequestMetricsResult{
		Range:        spec.name,
		Bucket:       spec.bucket,
		Total:        total,
		Success:      success,
		Failed:       failed,
		Series:       buckets,
		Environments: s.buildEnvBreakdown(ctx, userID, envCounts),
	}
	if total > 0 {
		result.AvgLatencyMs = latencySum / int64(total)
	}
	return result, nil
}

// buildEnvBreakdown joins per-environment counts with environment names and
// sorts them by request volume (descending). Environments with no traffic in
// the window are still listed so the UI can offer them in a picker.
func (s *MetricsService) buildEnvBreakdown(ctx context.Context, userID uuid.UUID, counts map[uuid.UUID]int) []RequestMetricsEnvBreakdown {
	envs, err := s.envs.ListByUserID(ctx, userID)
	if err != nil || len(envs) == 0 {
		return []RequestMetricsEnvBreakdown{}
	}
	out := make([]RequestMetricsEnvBreakdown, 0, len(envs))
	for _, env := range envs {
		out = append(out, RequestMetricsEnvBreakdown{
			ID:     env.ID.String(),
			Name:   env.Name,
			Count:  counts[env.ID],
			Active: env.Active,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// Record persists a single request metric. Recording failures are surfaced to
// the caller, which typically logs and ignores them so telemetry never blocks
// the user-facing response.
func (s *MetricsService) Record(ctx context.Context, userID uuid.UUID, envID *uuid.UUID, modelStr string, success bool, latency time.Duration) error {
	return s.requests.Create(ctx, &model.RequestMetric{
		UserID:        userID,
		EnvironmentID: envID,
		Model:         modelStr,
		Success:       success,
		LatencyMs:     latency.Milliseconds(),
		CreatedAt:     s.now().UTC(),
	})
}
