package domains

import (
	"time"

	"github.com/google/uuid"
)

type CollabStatus string

const (
	CollabPending  CollabStatus = "PENDING"
	CollabAccepted CollabStatus = "ACCEPTED"
	CollabRejected CollabStatus = "REJECTED"
)

type Collaboration struct {
	ID             uuid.UUID    `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	PatientID      uuid.UUID    `gorm:"type:uuid;not null"`
	ProfessionalID uuid.UUID    `gorm:"type:uuid;not null"` // El usuario invitado
	Status         CollabStatus `gorm:"type:varchar(20);default:'PENDING';not null"`

	InvitedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Relaciones para Preload
	Professional User    `gorm:"foreignKey:ProfessionalID"`
	Patient      Patient `gorm:"foreignKey:PatientID"`
}

// Input para invitar a alguien por email
type InviteInput struct {
	PatientID string `json:"patient_id" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}
