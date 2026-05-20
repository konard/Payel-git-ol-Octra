package accaunt

import "github.com/gin-gonic/gin"

func registerLogout(r *gin.Engine) {
	r.POST("/logout", func(c *gin.Context) {
		c.SetCookie("refresh_token", "", -1, "/", "", false, true)
		c.SetCookie("access_token", "", -1, "/", "", false, true)
		c.JSON(200, gin.H{"status": "ok", "message": "Logged out successfully"})
	})
}
