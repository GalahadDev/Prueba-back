package domains

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Definición de Roles
type UserRole string

const (
	RoleAdmin        UserRole = "ADMIN"
	RoleProfessional UserRole = "PROFESSIONAL"
	RoleBusiness     UserRole = "BUSINESS"
)

// Definición de Status
type UserStatus string

const (
	StatusInactive UserStatus = "INACTIVE"
	StatusActive   UserStatus = "ACTIVE"
	StatusRejected UserStatus = "REJECTED"
)

// User representa la tabla de usuarios en la BD
type User struct {
	ID     uuid.UUID  `gorm:"type:uuid;primary_key"`
	Email  string     `gorm:"uniqueIndex;not null"`
	Role   UserRole   `gorm:"type:user_role;default:'PROFESSIONAL'"`
	Status UserStatus `gorm:"type:user_status;default:'INACTIVE'"`

	// NUEVO CAMPO
	AvatarURL string `gorm:"type:text"`

	// AQUÍ SE GUARDARÁ TODO EL RESTO (Nombre completo, provider, etc.)
	ProfileData datatypes.JSON `gorm:"type:jsonb"`

	RejectReason string         `gorm:"type:text"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate asegura que el UUID se genere si no viene dado (aunque Postgres lo hace por default)
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
