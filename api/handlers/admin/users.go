package admin

import (
	"net/http"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
)

// ListPendingUsersHandler: Muestra quiénes quieren entrar
func ListPendingUsersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []domains.User
		// Filtramos solo los INACTIVE
		database.GetDB().Where("status = ?", domains.StatusInactive).Find(&users)
		c.JSON(http.StatusOK, gin.H{"data": users})
	}
}

// Input para revisar usuario
type ReviewUserInput struct {
	Action       string `json:"action" binding:"required,oneof=APPROVE REJECT"`
	RejectReason string `json:"reject_reason"` // Obligatorio si es REJECT
}

// ReviewUserHandler: El Martillo del Juez (Aprobar/Rechazar)
// AHORA RECIBE (cfg *config.Config) para poder enviar emails
func ReviewUserHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetUserID := c.Param("id")

		var input ReviewUserInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetDB()
		var user domains.User
		if err := db.First(&user, "id = ?", targetUserID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Lógica de Negocio
		switch input.Action {
		case "APPROVE":
			user.Status = domains.StatusActive
			user.RejectReason = "" // Limpiar si hubo rechazo previo
		case "REJECT":
			if input.RejectReason == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Reject reason is required"})
				return
			}
			user.Status = domains.StatusRejected
			user.RejectReason = input.RejectReason
		}

		if err := db.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
			return
		}

		// --- INTEGRACIÓN DE NOTIFICACIONES ---
		// Instanciamos el servicio con la configuración (SMTP) y enviamos la alerta
		notifier := services.NewNotificationService(cfg)
		notifier.NotifyAccountStatus(user.ID, user.Status, user.RejectReason)

		c.JSON(http.StatusOK, gin.H{
			"message":    "User status updated successfully",
			"new_status": user.Status,
		})
	}
}
