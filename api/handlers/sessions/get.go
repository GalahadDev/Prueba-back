package sessions

import (
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

// GetSessionHandler busca una sesión por su UUID
func GetSessionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var session domains.Session
		if err := database.GetDB().First(&session, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": session})
	}
}
