package router

import (
	"github.com/gin-gonic/gin"

	"user/internal/fetcher/http/router/accaunt"
	"user/internal/fetcher/http/router/app"
	"user/internal/fetcher/http/router/chat"
	"user/internal/fetcher/http/router/payments"
	"user/internal/fetcher/http/router/subscribe"
)

func RegisterAllRoutes(r *gin.Engine) {
	accaunt.RegisterRoutes(r)
	app.RegisterRoutes(r)
	chat.RegisterRoutes(r)
	payments.RegisterRoutes(r)
	subscribe.RegisterRoutes(r)
	RegisterHealthRoutes(r)
}
