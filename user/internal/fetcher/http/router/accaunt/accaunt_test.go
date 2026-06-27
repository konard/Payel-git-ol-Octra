package accaunt

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"

	"user/internal/core/services"
)

func TestStatusForRegisterError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"already exists", services.ErrUserAlreadyExists, 409},
		{"wrapped already exists", fmt.Errorf("register: %w", services.ErrUserAlreadyExists), 409},
		{"generic failure", errors.New("database unavailable"), 500},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := statusForRegisterError(tc.err); got != tc.want {
				t.Errorf("statusForRegisterError(%v) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}
}

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
