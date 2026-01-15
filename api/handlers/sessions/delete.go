package sessions

import (
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

func DeleteSessionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		currentUser := c.MustGet("currentUser").(domains.User)

		var session domains.Session
		db := database.GetDB()

		if err := db.First(&session, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		// Seguridad: Solo autor o Admin
		if session.ProfessionalID != currentUser.ID && currentUser.Role != domains.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this session"})
			return
		}

		// Soft Delete
		if err := db.Delete(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Session deleted successfully"})
	}
}
