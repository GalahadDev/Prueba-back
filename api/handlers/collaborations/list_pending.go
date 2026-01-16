package collaborations

import (
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

// GetPendingInvitationsHandler lista las invitaciones donde soy el profesional invitado
func GetPendingInvitationsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)

		var invitations []domains.Collaboration

		// Buscamos invitaciones donde:
		// 1. El ProfessionalID soy YO (el usuario logueado)
		// 2. El estado es PENDING
		// 3. Pre-cargamos los datos del Paciente para mostrar el nombre en la notificación
		if err := database.GetDB().
			Preload("Patient").
			Where("professional_id = ? AND status = ?", currentUser.ID, domains.CollabPending).
			Find(&invitations).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invitations"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": invitations})
	}
}
