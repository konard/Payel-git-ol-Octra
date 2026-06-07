package context

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"orchestrator/pkg/models"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	rdb     *redis.Client
	enabled bool
	ttl     time.Duration
}

func NewRedisCache(rdb *redis.Client, enabled bool) *RedisCache {
	return &RedisCache{
		rdb:     rdb,
		enabled: enabled,
		ttl:     RedisTTL,
	}
}

func (c *RedisCache) ctxKey(projectID string) string {
	return fmt.Sprintf("CTX:%s", projectID)
}

func (c *RedisCache) memberKey(entry *models.ContextEntry) string {
	switch entry.Scope {
	case "global":
		return fmt.Sprintf("global:%s", entry.ID)
	case "team":
		return fmt.Sprintf("team:%s:%s", entry.TargetID, entry.ID)
	case "individual":
		return fmt.Sprintf("indiv:%s:%s", entry.TargetID, entry.ID)
	default:
		return fmt.Sprintf("other:%s", entry.ID)
	}
}

func (c *RedisCache) Save(ctx context.Context, entry *models.ContextEntry) {
	if !c.enabled || c.rdb == nil {
		return
	}
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("[ContextCache] marshal error: %v", err)
		return
	}
	key := c.ctxKey(entry.ProjectID.String())
	member := c.memberKey(entry)

	if err := c.rdb.ZAdd(ctx, key, redis.Z{
		Score:  float64(entry.Timestamp.UnixNano()),
		Member: member,
	}).Err(); err != nil {
		log.Printf("[ContextCache] ZAdd error: %v", err)
		return
	}

	if err := c.rdb.Set(ctx, fmt.Sprintf("%s:data:%s", key, member), data, c.ttl).Err(); err != nil {
		log.Printf("[ContextCache] Set error: %v", err)
		return
	}

	c.rdb.Expire(ctx, key, c.ttl)
}

func (c *RedisCache) GetGlobal(ctx context.Context, projectID string) ([]models.ContextEntry, error) {
	return c.getByPrefix(ctx, projectID, "global:")
}

func (c *RedisCache) GetTeam(ctx context.Context, projectID, managerID string) ([]models.ContextEntry, error) {
	return c.getByPrefix(ctx, projectID, fmt.Sprintf("team:%s:", managerID))
}

func (c *RedisCache) GetIndividual(ctx context.Context, projectID, agentID string) ([]models.ContextEntry, error) {
	return c.getByPrefix(ctx, projectID, fmt.Sprintf("indiv:%s:", agentID))
}

func (c *RedisCache) getByPrefix(ctx context.Context, projectID, prefix string) ([]models.ContextEntry, error) {
	if !c.enabled || c.rdb == nil {
		return nil, nil
	}
	key := c.ctxKey(projectID)
	members, err := c.rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}).Result()
	if err != nil {
		return nil, err
	}

	var entries []models.ContextEntry
	for _, member := range members {
		if len(member) < len(prefix) || member[:len(prefix)] != prefix {
			continue
		}
		data, err := c.rdb.Get(ctx, fmt.Sprintf("%s:data:%s", key, member)).Bytes()
		if err != nil {
			continue
		}
		var entry models.ContextEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}
		if entry.Forgotten {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (c *RedisCache) Populate(ctx context.Context, entries []models.ContextEntry) {
	for i := range entries {
		c.Save(ctx, &entries[i])
	}
}

func (c *RedisCache) Delete(ctx context.Context, entry *models.ContextEntry) {
	if !c.enabled || c.rdb == nil {
		return
	}
	key := c.ctxKey(entry.ProjectID.String())
	member := c.memberKey(entry)
	c.rdb.ZRem(ctx, key, member)
	c.rdb.Del(ctx, fmt.Sprintf("%s:data:%s", key, member))
}

func (c *RedisCache) ClearProject(ctx context.Context, projectID string) {
	if !c.enabled || c.rdb == nil {
		return
	}
	c.rdb.Del(ctx, c.ctxKey(projectID))
}
