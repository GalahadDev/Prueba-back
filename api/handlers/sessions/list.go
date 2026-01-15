package sessions

import (
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

// ListSessionsHandler obtiene sesiones con filtros
func ListSessionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.GetDB()
		var sessions []domains.Session

		// Query base
		query := db.Model(&domains.Session{})

		// 1. Filtro por Paciente (El más común)
		patientID := c.Query("patient_id")
		if patientID != "" {
			query = query.Where("patient_id = ?", patientID)
		}

		// 2. Filtro por Profesional (Para reportes individuales)
		profID := c.Query("professional_id")
		if profID != "" {
			query = query.Where("professional_id = ?", profID)
		}

		// 3. Filtro por Incidentes (Para el dashboard de alertas)
		incident := c.Query("has_incident")
		if incident == "true" {
			query = query.Where("has_incident = ?", true)
		}

		// 4. Ordenamiento: Siempre lo más reciente primero
		// Usamos Preload para traer datos si tuvieramos relaciones definidas,
		// pero por ahora Session es self-contained.
		if err := query.Order("created_at DESC").Find(&sessions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sessions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": sessions})
	}
}
