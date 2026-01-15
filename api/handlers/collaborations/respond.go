package collaborations

import (
	"net/http"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
)

// Input para el JSON del body
type RespondInvitationInput struct {
	Status string `json:"status" binding:"required,oneof=ACCEPTED REJECTED"`
}

// RespondInvitationHandler ahora recibe la configuración para enviar correos
func RespondInvitationHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		collabID := c.Param("id")
		currentUser := c.MustGet("currentUser").(domains.User)

		var input RespondInvitationInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be ACCEPTED or REJECTED"})
			return
		}

		db := database.GetDB()
		var collab domains.Collaboration

		// 1. Buscar la invitación
		if err := db.Preload("Patient").First(&collab, "id = ?", collabID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
			return
		}

		// 2. SEGURIDAD: Verificar que SOY el invitado
		if collab.ProfessionalID != currentUser.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not the recipient of this invitation"})
			return
		}

		// 3. Lógica de Estado: Solo se responde si está PENDING
		if collab.Status != domains.CollabPending {
			c.JSON(http.StatusBadRequest, gin.H{"error": "This invitation has already been processed"})
			return
		}

		// 4. Actualizar Estado
		newStatus := domains.CollabStatus(input.Status)
		collab.Status = newStatus

		if err := db.Save(&collab).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invitation status"})
			return
		}

		// 5. NOTIFICACIÓN REAL: Avisar al Creador del Paciente [cite: 108]
		var creator domains.User
		// Buscamos al creador usando el CreatorID que viene en collab.Patient (gracias al Preload)
		if err := db.First(&creator, "id = ?", collab.Patient.CreatorID).Error; err == nil {
			notifier := services.NewNotificationService(cfg)
			notifier.NotifyInviteResponse(creator.ID, currentUser.Email, newStatus)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Invitation updated successfully",
			"status":  newStatus,
		})
	}
}
