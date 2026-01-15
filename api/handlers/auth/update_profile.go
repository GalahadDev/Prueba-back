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
}

func UpdateProfileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Obtener usuario autenticado (del contexto del middleware)
		currentUser := c.MustGet("currentUser").(domains.User)

		var input UpdateProfileInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetDB()

		// 2. Leer el profile_data actual (que tiene lo de Google)
		var currentProfile map[string]interface{}

		// Si está vacío o es nulo, iniciamos un mapa nuevo
		if len(currentUser.ProfileData) == 0 {
			currentProfile = make(map[string]interface{})
		} else {
			// Convertimos el JSONB de Gorm a un mapa de Go para editarlo
			if err := json.Unmarshal(currentUser.ProfileData, &currentProfile); err != nil {
				// Si falla, reseteamos para evitar errores
				currentProfile = make(map[string]interface{})
			}
		}

		// 3. Actualizar/Agregar los nuevos campos
		// Solo actualizamos si el usuario envió algo (no string vacío)
		if input.FullName != "" {
			currentProfile["full_name"] = input.FullName
		}
		if input.Specialty != "" {
			currentProfile["specialty"] = input.Specialty
		}
		if input.Phone != "" {
			currentProfile["phone"] = input.Phone
		}

		// 4. Volver a empaquetar en JSONB
		newProfileJSON, err := json.Marshal(currentProfile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process profile data"})
			return
		}

		// 5. Guardar en Base de Datos
		// Actualizamos solo el campo ProfileData para ser eficientes
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
