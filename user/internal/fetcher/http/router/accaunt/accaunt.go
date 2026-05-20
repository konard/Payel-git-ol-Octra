package accaunt

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
	registerRegister(r)
	registerLogin(r)
	registerOAuth(r)
	registerLogout(r)
	registerMe(r)
	registerRefresh(r)
}