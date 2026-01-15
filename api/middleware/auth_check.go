package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/domains"
	"bitacora-medica-backend/api/services"

	"github.com/MicahParks/keyfunc/v2" // Asegúrate que sea v2
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	// 1. Configuración JWKS (Solo se usará si el token NO es HS256)
	jwksURL := fmt.Sprintf("%s/auth/v1/.well-known/jwks.json", cfg.SupabaseURL)
	options := keyfunc.Options{
		RefreshInterval:   time.Hour,
		RefreshRateLimit:  time.Minute * 5,
		RefreshTimeout:    time.Second * 10,
		RefreshUnknownKID: true,
	}
	// Inicializamos JWKS pero no entramos en pánico si falla, porque tu llave actual es HS256
	jwks, err := keyfunc.Get(jwksURL, options)
	if err != nil {
		slog.Warn("JWKS init failed (It's OK if you are using Legacy HS256)", "error", err)
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. PARSEO INTELIGENTE (Soporta HS256 y JWKS)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

			// CASO A: Tu Supabase actual (HS256)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
				if cfg.JwtSecret == "" {
					return nil, fmt.Errorf("HS256 token received but JWT_SECRET is missing in .env")
				}
				return []byte(cfg.JwtSecret), nil
			}

			// CASO B: Si rotas las llaves en el futuro (RSA/ECDSA)
			if jwks != nil {
				return jwks.Keyfunc(token)
			}

			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		})

		if err != nil || !token.Valid {
			slog.Warn("Token validation failed", "error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token", "details": err.Error()})
			return
		}

		// 3. Extraer Claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token Claims"})
			return
		}

		supaUserIDStr, ok := claims["sub"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token missing 'sub' claim"})
			return
		}

		supaUserID, err := uuid.Parse(supaUserIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid UUID format"})
			return
		}

		// Email fallback
		userEmail, ok := claims["email"].(string)
		if !ok {
			userEmail = "no-email@provided"
		}

		// 4. LOGICA DE AUTO-REGISTRO
		var user domains.User
		db := database.GetDB()

		if err := db.Where("id = ?", supaUserID).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Extraer metadata
				metaDataMap := make(map[string]interface{})
				if metaClaim, ok := claims["user_metadata"].(map[string]interface{}); ok {
					metaDataMap = metaClaim
				}

				avatarURL := ""
				if val, ok := metaDataMap["avatar_url"].(string); ok {
					avatarURL = val
				} else if val, ok := metaDataMap["picture"].(string); ok {
					avatarURL = val
				}

				profileDataJSON, _ := json.Marshal(metaDataMap)

				newUser := domains.User{
					ID:          supaUserID,
					Email:       userEmail,
					Role:        domains.RoleProfessional,
					Status:      domains.StatusInactive,
					AvatarURL:   avatarURL,
					ProfileData: datatypes.JSON(profileDataJSON),
				}

				if createErr := db.Create(&newUser).Error; createErr != nil {
					slog.Error("Auto-register failed", "error", createErr)
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
					return
				}

				// Notificar
				slog.Info("New user auto-registered", "email", userEmail)
				notifier := services.NewNotificationService(cfg)
				notifier.NotifyNewUser(newUser.ID, newUser.Email)

				user = newUser
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}
		}

		// 5. Validar Status
		if user.Status == domains.StatusRejected {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Account REJECTED", "reason": user.RejectReason})
			return
		}

		if user.Status == domains.StatusInactive {
			if c.Request.Method == "PUT" && strings.Contains(c.Request.URL.Path, "/api/auth/profile") {
				c.Set("currentUser", user)
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Account Pending Approval"})
			return
		}

		c.Set("currentUser", user)
		c.Next()
	}
}
