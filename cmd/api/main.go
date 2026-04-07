package main

import (
	"log"
	"net/http"

	"grab/internal/core/config"
	"grab/internal/core/db"
	"grab/internal/modules/chat"
	"grab/internal/modules/earning"
	"grab/internal/modules/emergency"
	"grab/internal/modules/mood"
	"grab/internal/modules/reminder"
	"grab/internal/modules/sos"
	"grab/internal/modules/survivalmap"
	"grab/internal/modules/user"

	_ "grab/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Grab Driver Superapp API
// @version         1.0
// @description     Backend API cho Superapp dành cho Tài xế công nghệ.
// @description     Bao gồm: Auth, GPS Tracking, SOS, Chat, Map Sinh tồn, Mood, Earning, Reminder, Emergency Contact.

// @host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in   header
// @name Authorization
// @description  Nhập JWT token theo format: Bearer {token}

func main() {
	cfg := config.Load()

	pg := db.NewPostgres(cfg.PostgresDSN)
	rdb := db.NewRedis(cfg.RedisAddr, cfg.RedisPass)
	mongoClient := db.NewMongo(cfg.MongoURI)

	// Auto-migrate PostgreSQL tables
	pg.AutoMigrate(
		&user.User{},
		&earning.Trip{},
		&reminder.ReminderConfig{},
		&emergency.EmergencyContact{},
	) //nolint:errcheck
	survivalmap.NewRepository(pg).Migrate() //nolint:errcheck

	mux := http.NewServeMux()

	// --- Module routes ---
	userHandler := user.RegisterRoutes(mux, pg, rdb, cfg.JWTSecret)
	sos.RegisterRoutes(mux, pg, rdb, userHandler)
	survivalmap.RegisterRoutes(mux, pg, userHandler)
	chat.RegisterRoutes(mux, mongoClient, userHandler)
	mood.RegisterRoutes(mux, pg, mongoClient, userHandler)
	earning.RegisterRoutes(mux, pg, userHandler)
	reminder.RegisterRoutes(mux, pg, rdb, userHandler)
	emergency.RegisterRoutes(mux, pg, userHandler)

	// --- Swagger UI ---
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	log.Printf("[server] starting on :%s", cfg.Port)
	log.Printf("[server] Swagger UI → http://localhost:%s/swagger/index.html", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("[server] %v", err)
	}
}
