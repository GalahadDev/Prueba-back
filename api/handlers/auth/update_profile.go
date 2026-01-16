package auth

import (
	"encoding/json"
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// Estructura de lo que esperamos recibir del Frontend
type UpdateProfileInput struct {
	FullName  string `json:"full_name"`
	Specialty string `json:"specialty"`
	Phone     string `json:"phone"`
	Gender    string `json:"gender"`     // "Masculino", "Femenino", "Otro"
	Bio       string `json:"bio"`        // "Experto en kinesiología deportiva..."
	BirthDate string `json:"birth_date"` // YYYY-MM-DD
}

func UpdateProfileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)

		var input UpdateProfileInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetDB()

		// Leer el profile_data actual
		var currentProfile map[string]interface{}
		if len(currentUser.ProfileData) == 0 {
			currentProfile = make(map[string]interface{})
		} else {
			if err := json.Unmarshal(currentUser.ProfileData, &currentProfile); err != nil {
				currentProfile = make(map[string]interface{})
			}
		}

		// Actualizar/Agregar campos dinámicamente
		if input.FullName != "" {
			currentProfile["full_name"] = input.FullName
		}
		if input.Specialty != "" {
			currentProfile["specialty"] = input.Specialty
		}
		if input.Phone != "" {
			currentProfile["phone"] = input.Phone
		}
		// Nuevos campos al JSONB
		if input.Gender != "" {
			currentProfile["gender"] = input.Gender
		}
		if input.Bio != "" {
			currentProfile["bio"] = input.Bio
		}
		if input.BirthDate != "" {
			currentProfile["birth_date"] = input.BirthDate
		}

		// Empaquetar y Guardar
		newProfileJSON, err := json.Marshal(currentProfile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process profile data"})
			return
		}

		// Actualizamos profile_data
		if err := db.Model(&currentUser).Update("profile_data", datatypes.JSON(newProfileJSON)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Profile updated successfully",
			"data":    currentProfile,
		})
	}
}
