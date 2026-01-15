package sessions

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"bitacora-medica-backend/api/config" // Import necesario
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services" // Import necesario

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// CreateSessionHandler ahora requiere la configuración para enviar correos
func CreateSessionHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Obtener Profesional (Usuario Autenticado)
		currentUserInterface, _ := c.Get("currentUser")
		currentUser := currentUserInterface.(domains.User)

		// 2. Bind JSON
		var input domains.CreateSessionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 3. REGLA DE NEGOCIO: Validación de Incidentes
		if input.HasIncident && input.IncidentDetails == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Incident details are mandatory when an incident is reported.",
			})
			return
		}

		// 4. Procesar UUIDs y JSONB
		patientID, err := uuid.Parse(input.PatientID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Patient ID"})
			return
		}

		vitalsJSON, _ := json.Marshal(input.Vitals)

		// 5. Crear Modelo
		session := domains.Session{
			PatientID:          patientID,
			ProfessionalID:     currentUser.ID,
			InterventionPlan:   input.InterventionPlan,
			Vitals:             datatypes.JSON(vitalsJSON),
			Description:        input.Description,
			Achievements:       input.Achievements,
			PatientPerformance: input.PatientPerformance,
			Photos:             pq.StringArray(input.Photos),
			HasIncident:        input.HasIncident,
			IncidentDetails:    input.IncidentDetails,
			IncidentPhoto:      input.IncidentPhoto,
			NextSessionNotes:   input.NextSessionNotes,
		}

		// 6. Guardar en DB
		if err := database.DB.Create(&session).Error; err != nil {
			slog.Error("Failed to create session", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}

		// 7. REGLA DE NEGOCIO: Disparar Notificación de Incidente
		if session.HasIncident {
			// Instanciamos el servicio y notificamos al equipo
			notifier := services.NewNotificationService(cfg)
			notifier.NotifyIncident(session.PatientID, session.IncidentDetails)

			slog.Warn("INCIDENT REPORTED - Notifications triggered",
				"patient_id", session.PatientID,
				"professional", currentUser.Email)
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Session recorded successfully",
			"data":    session,
		})
	}
}
