package auth

import (
	"bitacora-medica-backend/api/domains"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetMeHandler devuelve los datos del usuario autenticado
func GetMeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Obtener usuario autenticado (del contexto del middleware)
		currentUser := c.MustGet("currentUser").(domains.User)

		// 2. Devolver los datos
		c.JSON(http.StatusOK, gin.H{
			"user": currentUser,
		})
	}
}
