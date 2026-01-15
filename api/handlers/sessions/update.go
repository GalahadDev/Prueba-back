package sessions

import (
	"encoding/json"
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// UpdateSessionHandler permite editar una sesión (Solo el autor)
func UpdateSessionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		currentUser := c.MustGet("currentUser").(domains.User)

		// 1. Buscar la sesión existente
		var session domains.Session
		db := database.GetDB()
		if err := db.First(&session, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		// 2. SEGURIDAD: Verificar que el usuario sea el autor
		if session.ProfessionalID != currentUser.ID && currentUser.Role != domains.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own sessions"})
			return
		}

		// 3. Bind de los nuevos datos
		var input domains.CreateSessionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 4. Actualizar campos
		// Nota: Convertimos los tipos complejos manualmente
		if input.Vitals != nil {
			vitalsJSON, _ := json.Marshal(input.Vitals)
			session.Vitals = datatypes.JSON(vitalsJSON)
		}

		session.InterventionPlan = input.InterventionPlan
		session.Description = input.Description
		session.Achievements = input.Achievements
		session.PatientPerformance = input.PatientPerformance
		session.NextSessionNotes = input.NextSessionNotes

		// Actualizar incidente si cambia
		session.HasIncident = input.HasIncident
		session.IncidentDetails = input.IncidentDetails
		session.IncidentPhoto = input.IncidentPhoto

		// Actualizar fotos si se envían nuevas
		if len(input.Photos) > 0 {
			session.Photos = pq.StringArray(input.Photos)
		}

		// 5. Guardar
		if err := db.Save(&session).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Session updated", "data": session})
	}
}
