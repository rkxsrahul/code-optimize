package api

import (
	"git.xenonstack.com/util/continuous-security-backend/src/web"
	"github.com/gin-gonic/gin"
)

func WorkspaceNameUpdate(c *gin.Context) {
	err := web.WorkspaceNameUpdate(c.Param("emails"))
	if err != nil {
		c.JSON(400, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"error":   false,
		"message": "workspace name update.",
	})

}
