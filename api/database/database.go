package database

import (
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(dbUrl string) {
	var err error

	// Configuración de GORM
	// PrepareStmt: true es más rápido, pero consume memoria en el servidor DB.
	// En Session Pool (5432) está bien usarlo. Si usaras Transaction Pool (6543), deberías poner false.
	gormConfig := &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Info), // Logs detallados para desarrollo
		PrepareStmt: true,
	}

	DB, err = gorm.Open(postgres.Open(dbUrl), gormConfig)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		panic("Failed to connect to database")
	}

	// --- OPTIMIZACIÓN SUPABASE FREE TIER ---
	sqlDB, err := DB.DB()
	if err != nil {
		slog.Error("Failed to get generic database object", "error", err)
		panic("Failed to configure connection pool")
	}

	// 1. MaxOpenConns: Límite estricto de conexiones activas.
	// Supabase Free ~60 total. Dejamos espacio para herramientas externas (TablePlus, Supabase Studio).
	// Usar 20 es seguro para una sola instancia del backend.
	sqlDB.SetMaxOpenConns(20)

	// 2. MaxIdleConns: Conexiones dormidas listas para reusar.
	// No mantener demasiadas abiertas sin uso para no saturar el pooler.
	sqlDB.SetMaxIdleConns(5)

	// 3. ConnMaxLifetime: Tiempo máximo que una conexión puede vivir.
	// Es bueno reciclar conexiones cada cierto tiempo en la nube para evitar timeouts fantasma.
	sqlDB.SetConnMaxLifetime(time.Hour)

	slog.Info("Database connection established successfully with pool limits")
}

func GetDB() *gorm.DB {
	return DB
}
