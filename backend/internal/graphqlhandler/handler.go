package graphqlhandler

import (
	"context"
	"encoding/json"
	"log"

	"backend/internal/model"

	"github.com/graphql-go/graphql"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

type AuthService interface {
	Authenticate(ctx context.Context, token string) (*model.User, error)
}

type DashboardEnvRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.DashboardEnvironment, error)
}

type CanvasNodeRepository interface {
	ListByEnvironment(ctx context.Context, envID uuid.UUID) ([]model.CanvasNode, error)
}

type ChatService interface {
	ChatWithEnvironment(ctx context.Context, user *model.User, envID uuid.UUID, prompt string) (string, error)
}

type Handler struct {
	schema graphql.Schema
}

func New(auth AuthService, envRepo DashboardEnvRepository, nodeRepo CanvasNodeRepository, chatSvc ChatService) (*Handler, error) {
	chatField := &graphql.Field{
		Type:        graphql.String,
		Description: "Send a chat prompt to an environment",
		Args: graphql.FieldConfigArgument{
			"environmentId": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"prompt": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"apiKey": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			envIDStr := p.Args["environmentId"].(string)
			prompt := p.Args["prompt"].(string)
			apiKey := p.Args["apiKey"].(string)

			envID, err := uuid.Parse(envIDStr)
			if err != nil {
				return nil, err
			}

			user, err := auth.Authenticate(context.Background(), apiKey)
			if err != nil {
				return nil, err
			}

			env, err := envRepo.GetByID(context.Background(), envID)
			if err != nil {
				return nil, err
			}
			if env.UserID != user.ID {
				return nil, nil
			}

			nodes, err := nodeRepo.ListByEnvironment(context.Background(), envID)
			if err != nil {
				return nil, err
			}
			if !hasAdapterNode(nodes, "graphql") {
				return nil, nil
			}

			return chatSvc.ChatWithEnvironment(context.Background(), user, envID, prompt)
		},
	}

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"chat": chatField,
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"ping": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return "pong", nil
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
	if err != nil {
		return nil, err
	}

	return &Handler{schema: schema}, nil
}

func hasAdapterNode(nodes []model.CanvasNode, protocol string) bool {
	for _, n := range nodes {
		if n.Kind != "adapter" {
			continue
		}
		var meta map[string]*string
		if n.Meta != "" {
			json.Unmarshal([]byte(n.Meta), &meta)
		}
		if meta != nil {
			if v := strPtr(meta["protocol"]); v == protocol {
				return true
			}
		}
	}
	return false
}

func strPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

type graphQLRequest struct {
	Query string `json:"query"`
}

func (h *Handler) ServeFastHTTP(ctx *fasthttp.RequestCtx) {
	var req graphQLRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(`{"error":"invalid json"}`)
		return
	}

	result := graphql.Do(graphql.Params{
		Schema:        h.schema,
		RequestString: req.Query,
	})

	ctx.SetContentType("application/json")
	if err := json.NewEncoder(ctx).Encode(result); err != nil {
		log.Printf("graphql encode: %v", err)
	}
}
