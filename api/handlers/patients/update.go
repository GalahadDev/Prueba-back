package patients

import (
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

type UpdatePatientInput struct {
	DisabilityReport string `json:"disability_report"`
	CareNotes        string `json:"care_notes"`
}

func UpdatePatientHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var input UpdatePatientInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetDB()
		var patient domains.Patient

		// 1. Buscar paciente
		if err := db.First(&patient, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
			return
		}

		// 2. Actualizar campos específicos
		// Nota: Solo actualizamos lo que necesitamos para no sobrescribir datos sensibles accidentalmente
		patient.DisabilityReport = input.DisabilityReport
		patient.CareNotes = input.CareNotes

		if err := db.Save(&patient).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update patient"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Patient updated successfully",
			"data":    patient,
		})
	}
}
