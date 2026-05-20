package app

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
	registerCustomModels(r)
	registerCustomProviders(r)
}