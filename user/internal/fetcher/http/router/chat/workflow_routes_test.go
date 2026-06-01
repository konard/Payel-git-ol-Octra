package chat

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesIncludesWorkflowLibraryRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r)

	routes := map[string]bool{}
	for _, route := range r.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	expected := []string{
		"POST /workflows",
		"GET /workflows/library",
		"GET /workflows/categories",
		"GET /workflows/my",
		"GET /workflows/:id",
		"POST /workflows/:id/download",
		"PUT /workflows/:id",
		"DELETE /workflows/:id",
	}
	for _, route := range expected {
		if !routes[route] {
			t.Fatalf("route %s is not registered; got %#v", route, routes)
		}
	}
}
