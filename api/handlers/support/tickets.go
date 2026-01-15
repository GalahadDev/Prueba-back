package support

import (
	"net/http"

	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/gin-gonic/gin"
)

// CreateTicketHandler (Para Usuarios)
func CreateTicketHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)
		var input domains.CreateTicketInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ticket := domains.SupportTicket{
			UserID:  currentUser.ID,
			Subject: input.Subject,
			Message: input.Message,
			Status:  domains.TicketOpen,
		}

		if err := database.GetDB().Create(&ticket).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Ticket created", "id": ticket.ID})
	}
}

// ListTicketsHandler (Dual: Admin ve todo, Usuario ve lo suyo)
func ListTicketsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser := c.MustGet("currentUser").(domains.User)
		db := database.GetDB()
		var tickets []domains.SupportTicket

		if currentUser.Role == domains.RoleAdmin {
			// Admin ve todo, ordenado por pendientes primero
			db.Preload("User").Order("status DESC, created_at ASC").Find(&tickets)
		} else {
			// Usuario ve solo lo suyo
			db.Where("user_id = ?", currentUser.ID).Order("created_at DESC").Find(&tickets)
		}

		c.JSON(http.StatusOK, gin.H{"data": tickets})
	}
}

// ReplyTicketHandler (Solo Admin)
func ReplyTicketHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketID := c.Param("id")
		var input domains.ReplyTicketInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetDB()
		var ticket domains.SupportTicket
		if err := db.First(&ticket, "id = ?", ticketID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
			return
		}

		// Actualizar respuesta y cerrar
		ticket.AdminResponse = input.Response
		ticket.Status = domains.TicketClosed

		if err := db.Save(&ticket).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reply ticket"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Ticket replied and closed"})
	}
}
