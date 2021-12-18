package api

import (
	"log"

	"git.xenonstack.com/util/continuous-security-backend/src/method"
	"git.xenonstack.com/util/continuous-security-backend/src/web"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
)

func Integration(c *gin.Context) {
	// extracting jwt claims
	claims := jwt.ExtractClaims(c)
	email, ok := claims["email"].(string)
	if !ok {
		log.Println("email not set")
		c.JSON(500, gin.H{"error": true, "message": "Please login again"})
		return
	}

	workspace := c.Query("workspace")
	if workspace == "" {
		workspace = method.ProjectNamebyEmail(email)
	}

	code, mapd := web.Integration(email, workspace, c.GetHeader("Authorization"))
	c.JSON(code, mapd)
}

func IntegrationbyID(c *gin.Context) {

	// extracting jwt claims
	claims := jwt.ExtractClaims(c)
	email, ok := claims["email"].(string)
	if !ok {
		log.Println("email not set")
		c.JSON(500, gin.H{"error": true, "message": "Please login again"})
		return
	}
	workspace := c.Query("workspace")
	if workspace == "" {
		workspace = method.ProjectNamebyEmail(email)
	}

	code, mapd := web.ScanInformation(c.Param("id"), workspace, email, c.GetHeader("Authorization"))
	c.JSON(code, mapd)
}
