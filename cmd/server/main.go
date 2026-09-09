package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/taufiqgit/news-api/docs"
	"github.com/taufiqgit/news-api/internal/config"
	"github.com/taufiqgit/news-api/internal/database"
	"github.com/taufiqgit/news-api/internal/router"
)

//	@title			News API
//	@version		1.0
//	@description	REST API untuk website berita. Mendukung auth JWT, artikel, kategori, tag, dan komentar.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	News API Support
//	@contact.email	support@news-api.dev

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Masukkan token dengan format: `Bearer <token>`

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Override Swagger host — ikut APP_URL supaya docs jalan di semua environment
	// (fix: Swagger UI di produksi tidak lagi manggil localhost dari browser user)
	host := strings.TrimPrefix(cfg.AppURL, "https://")
	host = strings.TrimPrefix(host, "http://")
	docs.SwaggerInfo.Host = strings.TrimSuffix(host, "/")
	docs.SwaggerInfo.BasePath = "/api/v1"
	if cfg.Server.Environment == "production" {
		docs.SwaggerInfo.Schemes = []string{"https"}
	}

	// Setup context dengan signal handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Connect ke database
	pool, err := database.New(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer pool.Close()
	log.Println("✅ database connected")

	// Setup router
	r := router.Setup(cfg, pool)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Jalankan server di goroutine
	go func() {
		log.Printf("🚀 server started on port %s (%s)", cfg.Server.Port, cfg.Server.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	log.Println("🛑 shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}
	log.Println("👋 server stopped")
}
