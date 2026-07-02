package service

import (
	"context"
	"testing"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"github.com/google/uuid"
)

// fixedClock returns a MetricsService whose clock is pinned to `at` so the
// bucket windows are deterministic across test runs.
func newMetricsAt(t *testing.T, at time.Time) (*MetricsService, repository.DashboardEnvironmentRepository) {
	t.Helper()
	db := newTestDB(t)
	requests := repository.NewRequestMetricsRepository(db)
	envs := repository.NewDashboardEnvironmentRepository(db)
	svc := NewMetricsService(requests, envs)
	svc.now = func() time.Time { return at }
	return svc, envs
}

func TestRequestMetricsAggregatesByDay(t *testing.T) {
	// A Wednesday at noon UTC; the 7d window covers 7 daily buckets ending today.
	now := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)
	svc, _ := newMetricsAt(t, now)
	ctx := context.Background()
	user := uuid.New()

	// Two successes today, one failure today, one success yesterday.
	svc.now = func() time.Time { return now }
	if err := svc.Record(ctx, user, nil, "", true, 100*time.Millisecond); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := svc.Record(ctx, user, nil, "", true, 300*time.Millisecond); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := svc.Record(ctx, user, nil, "", false, 200*time.Millisecond); err != nil {
		t.Fatalf("record: %v", err)
	}
	svc.now = func() time.Time { return now.Add(-24 * time.Hour) }
	if err := svc.Record(ctx, user, nil, "", true, 400*time.Millisecond); err != nil {
		t.Fatalf("record: %v", err)
	}
	svc.now = func() time.Time { return now }

	res, err := svc.RequestMetrics(ctx, user, nil, "7d")
	if err != nil {
		t.Fatalf("RequestMetrics: %v", err)
	}
	if res.Range != "7d" || res.Bucket != "day" {
		t.Fatalf("unexpected range/bucket: %s/%s", res.Range, res.Bucket)
	}
	if len(res.Series) != 7 {
		t.Fatalf("expected 7 daily buckets, got %d", len(res.Series))
	}
	if res.Total != 4 || res.Success != 3 || res.Failed != 1 {
		t.Fatalf("totals wrong: total=%d success=%d failed=%d", res.Total, res.Success, res.Failed)
	}
	// avg latency = (100+300+200+400)/4 = 250ms
	if res.AvgLatencyMs != 250 {
		t.Fatalf("expected avg latency 250, got %d", res.AvgLatencyMs)
	}

	// The last bucket is today: 3 requests (2 success, 1 failed).
	today := res.Series[len(res.Series)-1]
	if today.Count != 3 || today.Success != 2 || today.Failed != 1 {
		t.Fatalf("today bucket wrong: %+v", today)
	}
	// The bucket before it is yesterday: 1 request.
	yesterday := res.Series[len(res.Series)-2]
	if yesterday.Count != 1 || yesterday.Success != 1 {
		t.Fatalf("yesterday bucket wrong: %+v", yesterday)
	}
}

func TestRequestMetricsPerEnvironmentBreakdownAndScope(t *testing.T) {
	now := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)
	svc, envs := newMetricsAt(t, now)
	ctx := context.Background()
	user := uuid.New()

	envA := &model.DashboardEnvironment{UserID: user, Name: "alpha", Active: true}
	envB := &model.DashboardEnvironment{UserID: user, Name: "beta", Active: false}
	if err := envs.Create(ctx, envA); err != nil {
		t.Fatalf("create env: %v", err)
	}
	if err := envs.Create(ctx, envB); err != nil {
		t.Fatalf("create env: %v", err)
	}
	// GORM applies the column's default:true for the zero-value bool, so mark
	// beta inactive explicitly to exercise the Active flag in the breakdown.
	if err := envs.SetActive(ctx, envB.ID, false); err != nil {
		t.Fatalf("set active: %v", err)
	}

	// 2 requests to A, 1 to B.
	if err := svc.Record(ctx, user, &envA.ID, "", true, 0); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := svc.Record(ctx, user, &envA.ID, "", false, 0); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := svc.Record(ctx, user, &envB.ID, "", true, 0); err != nil {
		t.Fatalf("record: %v", err)
	}

	// Global view lists both environments, A first (higher count).
	res, err := svc.RequestMetrics(ctx, user, nil, "7d")
	if err != nil {
		t.Fatalf("RequestMetrics: %v", err)
	}
	if res.Total != 3 {
		t.Fatalf("expected 3 total, got %d", res.Total)
	}
	if len(res.Environments) != 2 {
		t.Fatalf("expected 2 env breakdowns, got %d", len(res.Environments))
	}
	if res.Environments[0].Name != "alpha" || res.Environments[0].Count != 2 {
		t.Fatalf("expected alpha with 2, got %+v", res.Environments[0])
	}
	if res.Environments[1].Name != "beta" || res.Environments[1].Count != 1 {
		t.Fatalf("expected beta with 1, got %+v", res.Environments[1])
	}
	if res.Environments[1].Active {
		t.Fatal("expected beta to be inactive")
	}

	// Scoped to env A only shows A's 2 requests.
	scoped, err := svc.RequestMetrics(ctx, user, &envA.ID, "7d")
	if err != nil {
		t.Fatalf("RequestMetrics scoped: %v", err)
	}
	if scoped.Total != 2 || scoped.Success != 1 || scoped.Failed != 1 {
		t.Fatalf("scoped totals wrong: %+v", scoped)
	}
}

func TestRequestMetricsHourlyRange(t *testing.T) {
	now := time.Date(2026, time.July, 1, 12, 30, 0, 0, time.UTC)
	svc, _ := newMetricsAt(t, now)
	ctx := context.Background()
	user := uuid.New()

	if err := svc.Record(ctx, user, nil, "", true, 0); err != nil {
		t.Fatalf("record: %v", err)
	}

	res, err := svc.RequestMetrics(ctx, user, nil, "24h")
	if err != nil {
		t.Fatalf("RequestMetrics: %v", err)
	}
	if res.Range != "24h" || res.Bucket != "hour" {
		t.Fatalf("unexpected range/bucket: %s/%s", res.Range, res.Bucket)
	}
	if len(res.Series) != 24 {
		t.Fatalf("expected 24 hourly buckets, got %d", len(res.Series))
	}
	if res.Total != 1 {
		t.Fatalf("expected 1 total, got %d", res.Total)
	}
	// The final bucket is the current hour and holds the single request.
	last := res.Series[len(res.Series)-1]
	if last.Count != 1 {
		t.Fatalf("expected current-hour bucket count 1, got %+v", last)
	}
}

func TestRequestMetricsEmptyWindow(t *testing.T) {
	now := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)
	svc, _ := newMetricsAt(t, now)
	ctx := context.Background()

	res, err := svc.RequestMetrics(ctx, uuid.New(), nil, "")
	if err != nil {
		t.Fatalf("RequestMetrics: %v", err)
	}
	if res.Total != 0 || res.AvgLatencyMs != 0 {
		t.Fatalf("expected empty result, got total=%d avg=%d", res.Total, res.AvgLatencyMs)
	}
	if len(res.Series) != 7 {
		t.Fatalf("expected 7 seeded buckets even when empty, got %d", len(res.Series))
	}
	if len(res.Environments) != 0 {
		t.Fatalf("expected no environments, got %d", len(res.Environments))
	}
}
