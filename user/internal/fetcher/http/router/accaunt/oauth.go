package accaunt

import (
	"github.com/gin-gonic/gin"
	"user/internal/fetcher/http/oauth"
)

func registerOAuth(r *gin.Engine) {
	oauth.RegisterGoogleRoutes(r)
	oauth.RegisterGithubRoutes(r)
	oauth.RegisterLeFineRoutes(r)
}
