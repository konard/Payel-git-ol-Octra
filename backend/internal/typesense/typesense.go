package typesense

import (
	"context"
	"fmt"
	"log"

	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
)

const SkillsCollection = "skills"

type ClientInterface interface {
	EnsureCollection(ctx context.Context) error
	IndexSkills(ctx context.Context, docs []SkillDocument) error
}

var _ ClientInterface = (*Client)(nil)

type SkillDocument struct {
	ID         string `json:"id"`
	SkillID    string `json:"skill_id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	InstallCmd string `json:"install_cmd"`
}

type Client struct {
	client *typesense.Client
}

func New(host, apiKey string) *Client {
	client := typesense.NewClient(
		typesense.WithServer(fmt.Sprintf("http://%s", host)),
		typesense.WithAPIKey(apiKey),
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
		DefaultSortingField: strPtr("name"),
	}

	_, err := c.client.Collection(SkillsCollection).Retrieve(ctx)
	if err != nil {
		_, err := c.client.Collections().Create(ctx, schema)
		if err != nil {
			return fmt.Errorf("typesense create collection: %w", err)
		}
		log.Printf("typesense: created collection %s", SkillsCollection)
	}
	return nil
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

func strPtr(s string) *string { return &s }
