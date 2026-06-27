package accaunt

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesRegistersAllAuthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r)

	routes := map[string]bool{}
	for _, route := range r.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	expected := []string{
		"POST /register",
		"POST /login",
		"POST /logout",
		"POST /refresh",
		"GET /me",
		"GET /auth/google",
		"GET /auth/google/callback",
		"GET /auth/github",
		"GET /auth/github/callback",
		"GET /auth/lefine",
		"GET /auth/lefine/callback",
	}
	for _, route := range expected {
		if !routes[route] {
			t.Fatalf("route %s is not registered; got %#v", route, routes)
		}
	}
}
