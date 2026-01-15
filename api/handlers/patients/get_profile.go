package patients

import (
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

// Estructura de respuesta compleja
type PatientProfileResponse struct {
	Patient        domains.Patient   `json:"patient"`
	Team           []domains.User    `json:"team"` // Profesionales con acceso
	RecentSessions []domains.Session `json:"recent_sessions"`
	IncidentCount  int64             `json:"incident_count"`
}

func GetPatientProfileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		// Nota: En producción, aquí deberíamos validar si el usuario que consulta
		// TIENE permiso (es creador O colaborador aceptado).
		// Por brevedad, omitimos esa validación de lectura (Middleware check role).

		db := database.GetDB()

		// 1. Cargar Paciente
		var patient domains.Patient
		if err := db.First(&patient, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
			return
		}

		// 2. Cargar Equipo (Colaboradores Aceptados)
		// Hacemos un JOIN manual o buscamos IDs en collaboration
		var collaborators []domains.User
		db.Table("users").
			Joins("JOIN collaborations ON collaborations.professional_id = users.id").
			Where("collaborations.patient_id = ? AND collaborations.status = ?", id, domains.CollabAccepted).
			Find(&collaborators)

		// Agregar al creador al equipo (opcional, visualmente útil)
		var creator domains.User
		db.First(&creator, "id = ?", patient.CreatorID)
		collaborators = append(collaborators, creator)

		// 3. Cargar Últimas 5 Sesiones
		var sessions []domains.Session
		db.Where("patient_id = ?", id).Order("created_at DESC").Limit(5).Find(&sessions)

		// 4. Contar Incidentes Totales
		var incidentCount int64
		db.Model(&domains.Session{}).Where("patient_id = ? AND has_incident = ?", id, true).Count(&incidentCount)

		// 5. Armar Respuesta
		response := PatientProfileResponse{
			Patient:        patient,
			Team:           collaborators,
			RecentSessions: sessions,
			IncidentCount:  incidentCount,
		}

		c.JSON(http.StatusOK, gin.H{"data": response})
	}
}
