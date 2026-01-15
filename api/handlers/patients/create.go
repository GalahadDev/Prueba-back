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

// Input Validado: Esto es lo que Postman envía.
// Las etiquetas `json:"..."` son OBLIGATORIAS para que Go entienda snake_case.
type CreatePatientInput struct {
	FirstName     string `json:"first_name" binding:"required"` // Postman envía "first_name"
	LastName      string `json:"last_name" binding:"required"`
	RUT           string `json:"rut" binding:"required"`        // Obligatorio
	BirthDate     string `json:"birth_date" binding:"required"` // YYYY-MM-DD
	Email         string `json:"email" binding:"required,email"`
	Phone         string `json:"phone"`
	Diagnosis     string `json:"diagnosis"`
	ConsentPDFUrl string `json:"consent_pdf_url" binding:"required"`
}

func CreatePatientHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)

		var input CreatePatientInput
		// 1. Gin intenta leer el JSON y valida los campos obligatorios
		if err := c.ShouldBindJSON(&input); err != nil {
			// Si falta "rut" o "first_name", aquí saltará el error que viste
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validar formato de fecha (opcional pero recomendado)
		if _, err := time.Parse("2006-01-02", input.BirthDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Birth date must be YYYY-MM-DD"})
			return
		}

		// 2. EMPAQUETADO: Convertimos los campos sueltos al JSONB `personal_info`
		// Esto es lo que se guardará en la columna 'personal_info' de la BD
		personalInfoMap := map[string]interface{}{
			"first_name": input.FirstName,
			"last_name":  input.LastName,
			"rut":        input.RUT,
			"birth_date": input.BirthDate,
			"email":      input.Email,
			"phone":      input.Phone,
			"diagnosis":  input.Diagnosis,
		}

		personalInfoBytes, err := json.Marshal(personalInfoMap)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process personal info"})
			return
		}

		// 3. Crear el Modelo para la BD
		patient := domains.Patient{
			CreatorID:     currentUser.ID,
			PersonalInfo:  datatypes.JSON(personalInfoBytes), // Aquí va el JSON empaquetado
			ConsentPDFUrl: input.ConsentPDFUrl,
			// DisabilityReport y CareNotes quedan vacíos al crear, se llenan luego si es necesario
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
