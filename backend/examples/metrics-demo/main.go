// Command metrics-demo boots a minimal server that exercises the REAL
// request-metrics stack (model.RequestMetric, repository.RequestMetricsRepository
// and service.MetricsService.RequestMetrics) so the dashboard charts can be
// rendered against representative data without needing a live LLM provider.
//
// The production write path (POST /api/chat -> API.recordChatMetric ->
// MetricsService.Record) is covered by backend/internal/api/api_test.go's
// TestRequestMetricsEndpoint. A real chat additionally requires a provisioned
// ocawe/Nix environment, which is unavailable in a sandbox, so this demo seeds
// a week of representative traffic straight through the real repository and then
// serves the exact same aggregation the backend exposes at
// GET /api/metrics/requests.
package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	ctx := context.Background()

	db, err := gorm.Open(sqlite.Open("file:metricsdemo?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(model.AllModels()...); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	requests := repository.NewRequestMetricsRepository(db)
	envs := repository.NewDashboardEnvironmentRepository(db)
	metrics := service.NewMetricsService(requests, envs)

	// A demo user that owns the seeded traffic.
	user := &model.User{ID: uuid.New(), Email: "demo@octra.dev"}
	if err := db.Create(user).Error; err != nil {
		log.Fatalf("create user: %v", err)
	}

	// Two dashboard environments: a busy production one and a quieter staging.
	prod := &model.DashboardEnvironment{ID: uuid.New(), UserID: user.ID, Name: "production-api", Active: true}
	staging := &model.DashboardEnvironment{ID: uuid.New(), UserID: user.ID, Name: "staging-api", Active: true}
	for _, e := range []*model.DashboardEnvironment{prod, staging} {
		if err := db.Create(e).Error; err != nil {
			log.Fatalf("create env: %v", err)
		}
	}
	// Staging is currently paused; SetActive is used because GORM applies the
	// Active column's `default:true` to a zero-value false on insert.
	if err := envs.SetActive(ctx, staging.ID, false); err != nil {
		log.Fatalf("pause staging: %v", err)
	}

	seedTraffic(ctx, requests, user.ID, prod.ID, staging.ID)

	log.Printf("seeded metrics demo (user=%s, prod=%s, staging=%s)", user.ID, prod.ID, staging.ID)

	handler := func(rc *fasthttp.RequestCtx) {
		// Permissive CORS so the statically-exported frontend can call us from
		// another origin during the demo.
		rc.Response.Header.Set("Access-Control-Allow-Origin", "*")
		rc.Response.Header.Set("Access-Control-Allow-Headers", "*")
		rc.Response.Header.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		if string(rc.Method()) == fasthttp.MethodOptions {
			rc.SetStatusCode(fasthttp.StatusNoContent)
			return
		}
		if string(rc.Path()) != "/api/metrics/requests" {
			rc.SetStatusCode(fasthttp.StatusNotFound)
			return
		}

		var envID *uuid.UUID
		if raw := string(rc.QueryArgs().Peek("env")); raw != "" {
			parsed, err := uuid.Parse(raw)
			if err != nil {
				rc.SetStatusCode(fasthttp.StatusBadRequest)
				return
			}
			envID = &parsed
		}

		result, err := metrics.RequestMetrics(rc, user.ID, envID, string(rc.QueryArgs().Peek("range")))
		if err != nil {
			rc.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}
		rc.SetContentType("application/json")
		_ = json.NewEncoder(rc).Encode(result)
	}

	addr := ":8080"
	log.Printf("metrics demo listening on %s (GET /api/metrics/requests)", addr)
	if err := fasthttp.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

// seedTraffic writes a week of representative request metrics straight through
// the real repository, mixing successes and failures across both environments
// plus some proxy-mode (no environment) chats.
func seedTraffic(ctx context.Context, requests repository.RequestMetricsRepository, userID, prodID, stagingID uuid.UUID) {
	now := time.Now().UTC()

	// Daily volume for the past 7 days: (day-offset, prod count, staging count).
	daily := []struct {
		dayAgo, prod, staging int
	}{
		{6, 34, 6},
		{5, 41, 9},
		{4, 28, 4},
		{3, 52, 11},
		{2, 47, 8},
		{1, 61, 13},
		{0, 39, 7},
	}

	write := func(envID *uuid.UUID, at time.Time, i int) {
		// ~12% failure rate, deterministic so the demo is reproducible.
		success := i%8 != 0
		latency := time.Duration(120+(i*37)%680) * time.Millisecond
		metric := &model.RequestMetric{
			UserID:        userID,
			EnvironmentID: envID,
			Model:         "openai/gpt-4o-mini",
			Success:       success,
			LatencyMs:     latency.Milliseconds(),
			CreatedAt:     at,
		}
		if err := requests.Create(ctx, metric); err != nil {
			log.Fatalf("seed metric: %v", err)
		}
	}

	for _, d := range daily {
		day := now.AddDate(0, 0, -d.dayAgo)
		for i := 0; i < d.prod; i++ {
			// Spread requests across the working hours of that day.
			at := time.Date(day.Year(), day.Month(), day.Day(), 8+(i%11), (i*7)%60, 0, 0, time.UTC)
			p := prodID
			write(&p, at, i)
		}
		for i := 0; i < d.staging; i++ {
			at := time.Date(day.Year(), day.Month(), day.Day(), 9+(i%9), (i*13)%60, 0, 0, time.UTC)
			s := stagingID
			write(&s, at, i)
		}
	}

	// A handful of proxy-mode chats not tied to a named environment.
	for i := 0; i < 12; i++ {
		at := now.Add(-time.Duration(i) * 90 * time.Minute)
		write(nil, at, i)
	}
}
