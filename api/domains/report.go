package domains

import (
	"time"

	"github.com/google/uuid"
)

// ProfessionalReport: El resumen periódico que escribe cada terapeuta
type ProfessionalReport struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	PatientID uuid.UUID `gorm:"type:uuid;not null"`
	AuthorID  uuid.UUID `gorm:"type:uuid;not null"`

	// Rango que cubre este reporte (Ej: "Marzo 2026")
	DateRangeStart time.Time `gorm:"type:date;not null"`
	DateRangeEnd   time.Time `gorm:"type:date;not null"`

	Content            string `gorm:"type:text;not null"` // Resumen cualitativo
	ObjectivesAchieved string `gorm:"type:text"`          // Objetivos logrados

	CreatedAt time.Time `gorm:"autoCreateTime"`

	// Relaciones
	Author User `gorm:"foreignKey:AuthorID"`
}

// Input para generar el Reporte Maestro (Filtros)
type MasterReportRequest struct {
	PatientID string `form:"patient_id" binding:"required"`
	StartDate string `form:"start_date" binding:"required" time_format:"2006-01-02"`
	EndDate   string `form:"end_date" binding:"required" time_format:"2006-01-02"`
}
