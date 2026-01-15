package reports

import (
	"net/http"
	"time"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

// Estructura de Salida (El Resumen Global)
type MasterReportResponse struct {
	GeneratedAt time.Time `json:"generated_at"`
	DateRange   string    `json:"date_range"`

	// Métricas Cuantitativas (Calculadas desde Sesiones)
	TotalSessions  int64 `json:"total_sessions"`
	TotalIncidents int64 `json:"total_incidents"`

	// Resumen por Área/Profesional (Agregado desde Reportes)
	ProfessionalSummaries []ProfessionalSummary `json:"professional_summaries"`
}

type ProfessionalSummary struct {
	ProfessionalName string `json:"professional_name"`
	Role             string `json:"role"` // Ej: Fonoaudiólogo
	Summary          string `json:"summary"`
	Objectives       string `json:"objectives"`
}

func GenerateMasterReportHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req domains.MasterReportRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetDB()

		// 1. Obtener Reportes Individuales en el rango
		var reports []domains.ProfessionalReport
		if err := db.Preload("Author").
			Where("patient_id = ? AND date_range_start >= ? AND date_range_end <= ?",
				req.PatientID, req.StartDate, req.EndDate).
			Find(&reports).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reports"})
			return
		}

		// 2. Calcular Métricas desde Sesiones (Hard Data)
		var totalSessions int64
		var totalIncidents int64

		// Count sesiones
		db.Model(&domains.Session{}).
			Where("patient_id = ? AND created_at BETWEEN ? AND ?", req.PatientID, req.StartDate, req.EndDate).
			Count(&totalSessions)

		// Count incidentes
		db.Model(&domains.Session{}).
			Where("patient_id = ? AND created_at BETWEEN ? AND ? AND has_incident = ?", req.PatientID, req.StartDate, req.EndDate, true).
			Count(&totalIncidents)

		// 3. Consolidar Información (Algoritmo de Agregación)
		var summaries []ProfessionalSummary

		for _, r := range reports {
			// Aquí extraemos lo valioso de cada experto
			// Podríamos concatenar si un mismo experto tiene 2 reportes en el periodo
			summaries = append(summaries, ProfessionalSummary{
				ProfessionalName: r.Author.Email, // Idealmente usar Name del ProfileData
				Role:             string(r.Author.Role),
				Summary:          r.Content,
				Objectives:       r.ObjectivesAchieved,
			})
		}

		// 4. Construir Respuesta Final
		response := MasterReportResponse{
			GeneratedAt:           time.Now(),
			DateRange:             req.StartDate + " to " + req.EndDate,
			TotalSessions:         totalSessions,
			TotalIncidents:        totalIncidents,
			ProfessionalSummaries: summaries,
		}

		c.JSON(http.StatusOK, gin.H{"data": response})
	}
}
