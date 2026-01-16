package patients

import (
	"encoding/json"
	"net/http"
	"time"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// Input Validado
type CreatePatientInput struct {
	FirstName      string `json:"first_name" binding:"required"`
	LastName       string `json:"last_name" binding:"required"`
	RUT            string `json:"rut" binding:"required"`
	BirthDate      string `json:"birth_date" binding:"required"` // YYYY-MM-DD
	Email          string `json:"email" binding:"required,email"`
	Phone          string `json:"phone"`
	Diagnosis      string `json:"diagnosis"`
	ConsentPDFUrl  string `json:"consent_pdf_url" binding:"required"`
	Sex            string `json:"sex" binding:"required"` // "Masculino", "Femenino"
	EmergencyPhone string `json:"emergency_phone"`
}

// Helper para calcular edad exacta
func calculateAge(birthDateStr string) int {
	birthDate, err := time.Parse("2006-01-02", birthDateStr)
	if err != nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - birthDate.Year()

	// Restar un año si aún no ha pasado el cumpleaños este año
	if now.YearDay() < birthDate.YearDay() {
		age--
	}
	return age
}

func CreatePatientHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)

		var input CreatePatientInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validar formato de fecha
		if _, err := time.Parse("2006-01-02", input.BirthDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Birth date must be YYYY-MM-DD"})
			return
		}

		// CALCULAR EDAD
		age := calculateAge(input.BirthDate)

		// EMPAQUETADO ACTUALIZADO
		personalInfoMap := map[string]interface{}{
			"first_name":      input.FirstName,
			"last_name":       input.LastName,
			"rut":             input.RUT,
			"birth_date":      input.BirthDate,
			"email":           input.Email,
			"phone":           input.Phone,
			"diagnosis":       input.Diagnosis,
			"sex":             input.Sex,
			"age":             age,
			"emergency_phone": input.EmergencyPhone,
		}

		personalInfoBytes, err := json.Marshal(personalInfoMap)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process personal info"})
			return
		}

		patient := domains.Patient{
			CreatorID:     currentUser.ID,
			PersonalInfo:  datatypes.JSON(personalInfoBytes),
			ConsentPDFUrl: input.ConsentPDFUrl,
		}

		if err := database.GetDB().Create(&patient).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create patient"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Patient created successfully",
			"data":    patient,
		})
	}
}
