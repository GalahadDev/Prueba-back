package common

import (
	"net/http"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/services"

	"github.com/gin-gonic/gin"
)

// UploadImageHandler maneja la subida de cualquier evidencia visual
func UploadImageHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Recibir archivo del form-data key "file"
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File is mandatory"})
			return
		}

		// 2. Subir usando el servicio
		storage := services.NewStorageService(cfg)
		url, err := storage.UploadImage(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image", "details": err.Error()})
			return
		}

		// 3. Devolver la URL para que el frontend la use
		c.JSON(http.StatusOK, gin.H{
			"message": "Image uploaded successfully",
			"url":     url,
		})
	}
}
