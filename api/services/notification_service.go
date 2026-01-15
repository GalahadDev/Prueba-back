package services

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"

	"github.com/google/uuid"
)

type NotificationService struct {
	cfg *config.Config
}

func NewNotificationService(cfg *config.Config) *NotificationService {
	return &NotificationService{cfg: cfg}
}

// --- CORE: ENVÍO REAL Y PERSISTENCIA ---

func (s *NotificationService) createAndNotify(userID uuid.UUID, notifType string, subject string, body string, relatedID *uuid.UUID) {
	db := database.GetDB()

	// 1. Guardar en Base de Datos (Requerimiento )
	notif := domains.Notification{
		UserID:    userID,
		Type:      notifType,
		Message:   subject + ": " + body, // Resumen para la UI
		RelatedID: relatedID,
		IsRead:    false,
	}

	// Usamos goroutine para que el guardado y envío no bloqueen al usuario
	go func() {
		// A. Guardar en DB
		if err := db.Create(&notif).Error; err != nil {
			slog.Error("Failed to save notification DB", "error", err)
		}

		// B. Obtener Email del destinatario
		var user domains.User
		if err := db.Select("email").First(&user, "id = ?", userID).Error; err != nil {
			slog.Error("User email not found for notification", "userID", userID)
			return
		}

		// C. Enviar Email Real
		s.sendRealEmail(user.Email, subject, body)
	}()
}

func (s *NotificationService) sendRealEmail(to string, subject string, body string) {
	// Configuración de autenticación SMTP
	auth := smtp.PlainAuth("", s.cfg.SMTPEmail, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	// Cabeceras estándar de email
	headers := []string{
		"To: " + to,
		"From: Bitácora Médica <" + s.cfg.SMTPEmail + ">",
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"UTF-8\"",
		"\r\n",
	}

	msg := []byte(strings.Join(headers, "") + body)

	// Envío
	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)
	err := smtp.SendMail(addr, auth, s.cfg.SMTPEmail, []string{to}, msg)

	if err != nil {
		slog.Error("❌ FAILED to send email", "to", to, "error", err)
	} else {
		slog.Info("✅ Email sent successfully", "to", to, "subject", subject)
	}
}

// --- MÉTODOS DE NEGOCIO (Los 5 Eventos del PDF) ---

// 1. NewUser[cite: 104]: Avisar a todos los ADMINs
func (s *NotificationService) NotifyNewUser(newUserID uuid.UUID, userEmail string) {
	var admins []domains.User
	database.GetDB().Where("role = ?", domains.RoleAdmin).Find(&admins)

	subject := "Nuevo Usuario Registrado"
	body := fmt.Sprintf("El usuario %s se ha registrado y espera verificación. Por favor ingrese al panel administrativo.", userEmail)

	for _, admin := range admins {
		// Notificar a cada admin
		s.createAndNotify(admin.ID, "NEW_USER", subject, body, &newUserID)
	}
}

// 2. AccountStatus[cite: 105]: Aprobación o Rechazo
func (s *NotificationService) NotifyAccountStatus(userID uuid.UUID, status domains.UserStatus, reason string) {
	subject := "Actualización de Estado de Cuenta"
	body := fmt.Sprintf("Su cuenta ha sido: %s.", status)

	if status == domains.StatusRejected {
		body += fmt.Sprintf("\n\nMotivo del rechazo: %s", reason)
	} else {
		body += "\n\nYa puede acceder a la plataforma y gestionar sus pacientes."
	}

	s.createAndNotify(userID, "ACCOUNT_STATUS", subject, body, nil)
}

// 3. IncidentAlert[cite: 106]: A todo el equipo
func (s *NotificationService) NotifyIncident(patientID uuid.UUID, incidentDetails string) {
	db := database.GetDB()

	// Obtener Paciente (para el nombre)
	var patient domains.Patient
	db.First(&patient, "id = ?", patientID)
	// Como 'PersonalInfo' es JSON, aquí asumimos un string simple para el ejemplo,
	// en producción parsearías el JSON para sacar "Juan Perez".
	patientName := "Paciente ID " + patientID.String()

	// 1. Buscar colaboradores ACEPTADOS
	var collaborators []domains.User
	db.Table("users").
		Joins("JOIN collaborations ON collaborations.professional_id = users.id").
		Where("collaborations.patient_id = ? AND collaborations.status = ?", patientID, domains.CollabAccepted).
		Find(&collaborators)

	// 2. Buscar al Creador
	var creator domains.User
	db.First(&creator, "id = ?", patient.CreatorID)

	recipients := append(collaborators, creator)

	// Usamos un mapa para evitar duplicados si el creador también está en collaborations (raro pero posible)
	uniqueUsers := make(map[string]domains.User)
	for _, u := range recipients {
		uniqueUsers[u.ID.String()] = u
	}

	subject := "⚠️ ALERTA DE INCIDENTE: " + patientName
	body := fmt.Sprintf("Se ha reportado un incidente para el paciente %s.\n\nDetalle: %s\n\nPor favor revise la bitácora para más detalles y evidencia.", patientName, incidentDetails)

	for _, professional := range uniqueUsers {
		s.createAndNotify(professional.ID, "INCIDENT_ALERT", subject, body, &patientID)
	}
}

// 4. CollabInvite[cite: 107]: Invitación recibida
func (s *NotificationService) NotifyCollabInvite(invitedUserID uuid.UUID, patientID uuid.UUID) {
	subject := "Invitación a Colaborar"
	body := "Has sido invitado a colaborar en el expediente clínico de un paciente. Ingresa a la app para aceptar o rechazar."

	s.createAndNotify(invitedUserID, "COLLAB_INVITE", subject, body, &patientID)
}

// 5. InviteResponse[cite: 108]: Aviso al creador
func (s *NotificationService) NotifyInviteResponse(creatorID uuid.UUID, responderEmail string, status domains.CollabStatus) {
	subject := fmt.Sprintf("Invitación %s", status)
	body := fmt.Sprintf("El profesional %s ha respondido a tu invitación: %s.", responderEmail, status)

	s.createAndNotify(creatorID, "INVITE_RESPONSE", subject, body, nil)
}
