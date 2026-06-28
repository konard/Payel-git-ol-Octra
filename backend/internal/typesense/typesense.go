package typesense

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
)

const (
	SkillsCollection    = "skills"
	CLIsCollection      = "clis"
	ProvidersCollection = "providers"
)

type ClientInterface interface {
	EnsureCollection(ctx context.Context) error
	IndexSkills(ctx context.Context, docs []SkillDocument) error
	IndexCLIs(ctx context.Context, docs []CLIDocument) error
}

var _ ClientInterface = (*Client)(nil)

type SkillDocument struct {
	ID         string `json:"id"`
	SkillID    string `json:"skill_id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	InstallCmd string `json:"install_cmd"`
}

type CLIDocument struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	NixAttr    string `json:"nix_attr"`
	InstallCmd string `json:"install_cmd"`
}

type ProviderDocument struct {
	ID           string `json:"id"`
	Key          string `json:"key"`
	Name         string `json:"name"`
	BaseURL      string `json:"base_url"`
	AuthEnv      string `json:"auth_env"`
	DefaultModel string `json:"default_model"`
	Description  string `json:"description"`
}

type Client struct {
	client *typesense.Client
}

func New(host, apiKey string) *Client {
	client := typesense.NewClient(
		typesense.WithServer(fmt.Sprintf("http://%s", host)),
		typesense.WithAPIKey(apiKey),
		typesense.WithConnectionTimeout(5*time.Second),
	)
	return &Client{client: client}
}

func (c *Client) EnsureCollection(ctx context.Context) error {
	schema := &api.CollectionSchema{
		Name: SkillsCollection,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "skill_id", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "source", Type: "string"},
			{Name: "install_cmd", Type: "string"},
		},
	}

	var lastErr error
	for i := 0; i < 12; i++ {
		_, err := c.client.Collection(SkillsCollection).Retrieve(ctx)
		if err == nil {
			return nil
		}
		lastErr = err

		if strings.Contains(err.Error(), "Not Ready or Lagging") {
			log.Printf("typesense: not ready yet, retrying in 5s (attempt %d/12)", i+1)
			time.Sleep(5 * time.Second)
			continue
		}

		_, err = c.client.Collections().Create(ctx, schema)
		if err != nil {
			if strings.Contains(err.Error(), "Not Ready or Lagging") {
				log.Printf("typesense: not ready yet, retrying in 5s (attempt %d/12)", i+1)
				time.Sleep(5 * time.Second)
				continue
			}
			return fmt.Errorf("typesense create collection: %w", err)
		}
		log.Printf("typesense: created collection %s", SkillsCollection)
		return nil
	}

	return fmt.Errorf("typesense not ready after 12 retries: %w", lastErr)
}

func (c *Client) IndexSkill(ctx context.Context, doc SkillDocument) error {
	_, err := c.client.Collection(SkillsCollection).Documents().Upsert(ctx, doc)
	if err != nil {
		return fmt.Errorf("typesense index: %w", err)
	}
	return nil
}

func (c *Client) IndexSkills(ctx context.Context, docs []SkillDocument) error {
	if len(docs) == 0 {
		return nil
	}
	interfaces := make([]interface{}, len(docs))
	for i, d := range docs {
		interfaces[i] = d
	}
	action := "upsert"
	params := &api.ImportDocumentsParams{Action: &action}
	_, err := c.client.Collection(SkillsCollection).Documents().Import(ctx, interfaces, params)
	if err != nil {
		return fmt.Errorf("typesense import: %w", err)
	}
	return nil
}

func (c *Client) SearchSkills(ctx context.Context, query string, limit int) (*api.SearchResult, error) {
	sortBy := "_text_match:desc"
	params := &api.SearchCollectionParams{
		Q:       query,
		QueryBy: "name,skill_id,source",
		SortBy:  &sortBy,
		PerPage: &limit,
	}
	result, err := c.client.Collection(SkillsCollection).Documents().Search(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("typesense search: %w", err)
	}
	return result, nil
}

// --- CLI collection ---------------------------------------------------------

func (c *Client) EnsureCLICollection(ctx context.Context) error {
	schema := &api.CollectionSchema{
		Name: CLIsCollection,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "nix_attr", Type: "string"},
			{Name: "install_cmd", Type: "string"},
		},
	}
	var lastErr error
	for i := 0; i < 12; i++ {
		_, err := c.client.Collection(CLIsCollection).Retrieve(ctx)
		if err == nil {
			return nil
		}
		lastErr = err
		if strings.Contains(err.Error(), "Not Ready or Lagging") {
			log.Printf("typesense: not ready yet, retrying in 5s (attempt %d/12)", i+1)
			time.Sleep(5 * time.Second)
			continue
		}
		_, err = c.client.Collections().Create(ctx, schema)
		if err != nil {
			if strings.Contains(err.Error(), "Not Ready or Lagging") {
				log.Printf("typesense: not ready yet, retrying in 5s (attempt %d/12)", i+1)
				time.Sleep(5 * time.Second)
				continue
			}
			return fmt.Errorf("typesense create cli collection: %w", err)
		}
		log.Printf("typesense: created collection %s", CLIsCollection)
		return nil
	}
	return fmt.Errorf("typesense not ready after 12 retries: %w", lastErr)
}

func (c *Client) IndexCLIs(ctx context.Context, docs []CLIDocument) error {
	if len(docs) == 0 {
		return nil
	}
	interfaces := make([]interface{}, len(docs))
	for i, d := range docs {
		interfaces[i] = d
	}
	action := "upsert"
	params := &api.ImportDocumentsParams{Action: &action}
	_, err := c.client.Collection(CLIsCollection).Documents().Import(ctx, interfaces, params)
	if err != nil {
		return fmt.Errorf("typesense import clis: %w", err)
	}
	return nil
}

func (c *Client) SearchCLIs(ctx context.Context, query string, limit int) (*api.SearchResult, error) {
	sortBy := "_text_match:desc"
	params := &api.SearchCollectionParams{
		Q:       query,
		QueryBy: "name",
		SortBy:  &sortBy,
		PerPage: &limit,
	}
	result, err := c.client.Collection(CLIsCollection).Documents().Search(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("typesense search clis: %w", err)
	}
	return result, nil
}

// --- Provider collection ----------------------------------------------------

func (c *Client) EnsureProviderCollection(ctx context.Context) error {
	schema := &api.CollectionSchema{
		Name: ProvidersCollection,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "key", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "base_url", Type: "string"},
			{Name: "auth_env", Type: "string"},
			{Name: "default_model", Type: "string"},
			{Name: "description", Type: "string"},
		},
	}
	var lastErr error
	for i := 0; i < 12; i++ {
		_, err := c.client.Collection(ProvidersCollection).Retrieve(ctx)
		if err == nil {
			return nil
		}
		lastErr = err
		if strings.Contains(err.Error(), "Not Ready or Lagging") {
			log.Printf("typesense: not ready yet, retrying in 5s (attempt %d/12)", i+1)
			time.Sleep(5 * time.Second)
			continue
		}
		_, err = c.client.Collections().Create(ctx, schema)
		if err != nil {
			if strings.Contains(err.Error(), "Not Ready or Lagging") {
				log.Printf("typesense: not ready yet, retrying in 5s (attempt %d/12)", i+1)
				time.Sleep(5 * time.Second)
				continue
			}
			return fmt.Errorf("typesense create provider collection: %w", err)
		}
		log.Printf("typesense: created collection %s", ProvidersCollection)
		return nil
	}
	return fmt.Errorf("typesense not ready after 12 retries: %w", lastErr)
}

func (c *Client) IndexProviders(ctx context.Context, docs []ProviderDocument) error {
	if len(docs) == 0 {
		return nil
	}
	interfaces := make([]interface{}, len(docs))
	for i, d := range docs {
		interfaces[i] = d
	}
	action := "upsert"
	params := &api.ImportDocumentsParams{Action: &action}
	_, err := c.client.Collection(ProvidersCollection).Documents().Import(ctx, interfaces, params)
	if err != nil {
		return fmt.Errorf("typesense import providers: %w", err)
	}
	return nil
}

func (c *Client) SearchProviders(ctx context.Context, query string, limit int) (*api.SearchResult, error) {
	sortBy := "_text_match:desc"
	params := &api.SearchCollectionParams{
		Q:       query,
		QueryBy: "name,key,base_url,description",
		SortBy:  &sortBy,
		PerPage: &limit,
	}
	result, err := c.client.Collection(ProvidersCollection).Documents().Search(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("typesense search providers: %w", err)
	}
	return result, nil
}

func strPtr(s string) *string { return &s }
