package main

import (
	"log/slog"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/handlers/admin"
	"bitacora-medica-backend/api/handlers/auth"
	"bitacora-medica-backend/api/handlers/collaborations"
	"bitacora-medica-backend/api/handlers/common"
	"bitacora-medica-backend/api/handlers/patients"
	"bitacora-medica-backend/api/handlers/reports"
	"bitacora-medica-backend/api/handlers/sessions"
	"bitacora-medica-backend/api/handlers/support"
	"time"

	"github.com/gin-contrib/cors"

	//"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Cargar Configuración
	cfg := config.LoadConfig()

	// 2. Conectar a BD
	database.Connect(cfg.DBUrl)

	// 4. Configurar Router
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://tradelog-app.vercel.app", "https://cron-job.org", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Aumentar el límite de memoria para subida de archivos (ej: 8MB) si es necesario
	r.MaxMultipartMemory = 8 << 20

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	api := r.Group("/api")

	// Pasamos 'cfg' al middleware para validar JWT
	api.Use(middleware.AuthMiddleware(cfg))
	{
		authGroup := api.Group("/auth")
		{
			authGroup.PUT("/profile", auth.UpdateProfileHandler())
			authGroup.GET("/me", auth.GetMeHandler())
		}

		// --- GRUPO DE PACIENTES ---
		patientsGroup := api.Group("/patients")
		{
			patientsGroup.POST("/", patients.CreatePatientHandler(cfg))

			patientsGroup.GET("/", patients.ListPatientsHandler())

			// NUEVO: Perfil Unificado (Ojo de Dios del Paciente)
			patientsGroup.GET("/:id", patients.GetPatientProfileHandler())
		}

		sessionsGroup := api.Group("/sessions")
		{
			// CREATE
			sessionsGroup.POST("/", sessions.CreateSessionHandler(cfg))

			// READ (Listar con filtros: ?patient_id=...&has_incident=true)
			sessionsGroup.GET("/", sessions.ListSessionsHandler())

			// READ ONE (Detalle específico)
			sessionsGroup.GET("/:id", sessions.GetSessionHandler())

			// UPDATE (Solo autor)
			sessionsGroup.PUT("/:id", sessions.UpdateSessionHandler())

			// DELETE (Solo autor - Soft Delete)
			sessionsGroup.DELETE("/:id", sessions.DeleteSessionHandler())
		}

		uploads := api.Group("/uploads")
		uploads.POST("/image", common.UploadImageHandler(cfg))
	}

	collabGroup := api.Group("/collaborations")
	{
		// Invitar: POST /api/collaborations/invite
		collabGroup.POST("/invite", collaborations.InviteCollabHandler(cfg))

		// Responder: PUT /api/collaborations/:id/respond
		// :id es el ID de la COLABORACIÓN (no del paciente ni usuario)
		collabGroup.PUT("/:id/respond", collaborations.RespondInvitationHandler(cfg))
	}

	// --- GRUPO REPORTES ---
	reportsGroup := api.Group("/reports")
	{
		// Individual: POST /api/reports/ (Kine sube su resumen mensual)
		reportsGroup.POST("/", reports.CreateIndividualReportHandler())

		// Maestro: GET /api/reports/master?patient_id=...&start_date=...&end_date=...
		// (Admin/Dueño obtiene la visión global)
		reportsGroup.GET("/master", reports.GenerateMasterReportHandler())
	}

	// --- GRUPO SOPORTE (Accesible para todos) ---
	supportGroup := api.Group("/support")
	{
		supportGroup.POST("/", support.CreateTicketHandler())
		supportGroup.GET("/", support.ListTicketsHandler()) // Admin ve todo, User ve suyo

		// Responder ticket (Solo Admin)
		supportGroup.PUT("/:id/reply", middleware.RequireAdmin(), support.ReplyTicketHandler())
	}

	// --- GRUPO ADMIN (Protegido por RequireAdmin) ---
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.RequireAdmin())
	{
		// Gestión de Usuarios
		adminGroup.GET("/users/pending", admin.ListPendingUsersHandler())
		adminGroup.PUT("/users/:id/review", admin.ReviewUserHandler(cfg))

		// Dashboard (KPIs simples)
		adminGroup.GET("/dashboard", func(c *gin.Context) {
			// Implementación rápida de KPIs [cite: 113]
			var totalUsers, activePatients, incidentsToday int64
			db := database.GetDB()
			db.Model(&domains.User{}).Count(&totalUsers)
			db.Model(&domains.Patient{}).Count(&activePatients)
			db.Model(&domains.Session{}).Where("has_incident = ?", true).Count(&incidentsToday)

			c.JSON(200, gin.H{
				"total_users":        totalUsers,
				"active_patients":    activePatients,
				"incidents_all_time": incidentsToday,
			})
		})
	}

	slog.Info("Server starting on port " + cfg.Port)
	r.Run(":" + cfg.Port)
}
