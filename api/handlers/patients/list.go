package patients

import (
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

// ListPatientsHandler devuelve la lista de pacientes
// 1. Creados por el profesional actual
// 2. O compartidos con él mediante una colaboración ACEPTADA
func ListPatientsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)
		var patients []domains.Patient

		db := database.GetDB()

		// Consulta: (creator_id = yo) OR (id IN subquery_colabs_aceptadas)
		// subquery: select patient_id from collaborations where professional_id = yo AND status = 'ACCEPTED'

		err := db.Where("creator_id = ?", currentUser.ID).
			Or("id IN (?)", db.Table("collaborations").
				Select("patient_id").
				Where("professional_id = ? AND status = ?", currentUser.ID, domains.CollabAccepted)).
			Find(&patients).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch patients"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": patients,
		})
	}
}
