package collaborations

import (
	"net/http"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
)

// InviteCollabHandler ahora recibe la configuración para enviar correos
func InviteCollabHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)
		var input domains.InviteInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetDB()

		// 1. Validar que el paciente existe y YO soy el creador
		var patient domains.Patient
		if err := db.Where("id = ? AND creator_id = ?", input.PatientID, currentUser.ID).First(&patient).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Patient not found or you are not the creator"})
			return
		}

		// 2. Buscar al profesional invitado por email
		var invitedUser domains.User
		if err := db.Where("email = ?", input.Email).First(&invitedUser).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User with this email not found in the platform"})
			return
		}

		// 3. Evitar auto-invitación
		if invitedUser.ID == currentUser.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot invite yourself"})
			return
		}

		// 4. Crear Colaboración
		collab := domains.Collaboration{
			PatientID:      patient.ID,
			ProfessionalID: invitedUser.ID,
			Status:         domains.CollabPending,
		}

		// Usamos FirstOrCreate para evitar duplicar invitaciones
		if err := db.Where("patient_id = ? AND professional_id = ?", patient.ID, invitedUser.ID).FirstOrCreate(&collab).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invitation"})
			return
		}

		// 5. NOTIFICACIÓN REAL: Avisar al invitado por correo [cite: 107]
		notifier := services.NewNotificationService(cfg)
		notifier.NotifyCollabInvite(invitedUser.ID, patient.ID)

		c.JSON(http.StatusCreated, gin.H{"message": "Invitation sent", "data": collab})
	}
}
