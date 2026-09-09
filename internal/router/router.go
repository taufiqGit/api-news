package router

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/taufiqgit/news-api/internal/config"
	"github.com/taufiqgit/news-api/internal/handler"
	"github.com/taufiqgit/news-api/internal/middleware"
	"github.com/taufiqgit/news-api/internal/repository"
	"github.com/taufiqgit/news-api/internal/storage"
	"github.com/taufiqgit/news-api/internal/usecase"

	_ "github.com/taufiqgit/news-api/docs"
)

// uploadDir mengembalikan jalur absolut folder uploads
func uploadDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "uploads"
	}
	return filepath.Join(wd, "uploads")
}

// Setup membangun seluruh router dengan dependency injection
func Setup(cfg *config.Config, pool *pgxpool.Pool) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// Health check
	r.GET("/healthz", handler.HealthCheck)

	// Serve file upload (local storage fallback — dev mode)
	// Gunakan jalur absolut supaya bekerja dari mana pun server dijalankan
	r.Static("/uploads", filepath.Join(uploadDir()))

	// Swagger UI (docs API)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// --- Repositories (data layer) ---
	userRepo := repository.NewUserRepository(pool)
	categoryRepo := repository.NewCategoryRepository(pool)
	tagRepo := repository.NewTagRepository(pool)
	articleRepo := repository.NewArticleRepository(pool)
	commentRepo := repository.NewCommentRepository(pool)
	websiteRepo := repository.NewWebsiteRepository(pool)

	// --- Usecases (business layer) ---
	tagUsecase := usecase.NewTagUsecase(tagRepo)
	userUsecase := usecase.NewUserUsecase(userRepo)
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo)
	articleUsecase := usecase.NewArticleUsecase(articleRepo, tagUsecase)
	commentUsecase := usecase.NewCommentUsecase(commentRepo, articleRepo)
	websiteUsecase := usecase.NewWebsiteUsecase(websiteRepo)

	// --- Middleware / Auth ---
	tokenMgr := middleware.NewTokenManager(cfg.JWT)
	auth := tokenMgr.AuthMiddleware()
	requireAdmin := middleware.RoleMiddleware("admin")
	requireEditor := middleware.RoleMiddleware("admin", "editor")
	requireAuthor := middleware.RoleMiddleware("admin", "editor", "author")

	// --- Storage (S3) ---
	store, err := storage.New(cfg.S3)
	if err != nil {
		panic("failed to init storage: " + err.Error())
	}

	// --- Handlers (presentation layer) ---
	userHandler := handler.NewUserHandler(userUsecase, tokenMgr)
	categoryHandler := handler.NewCategoryHandler(categoryUsecase)
	tagHandler := handler.NewTagHandler(tagUsecase)
	articleHandler := handler.NewArticleHandler(articleUsecase)
	commentHandler := handler.NewCommentHandler(commentUsecase)
	uploadHandler := handler.NewUploadHandler(store)
	websiteHandler := handler.NewWebsiteHandler(websiteUsecase)

	// --- Routes ---
	api := r.Group("/api/v1")

	// Auth
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", userHandler.Register)
		authGroup.POST("/login", userHandler.Login)
	}

	// Users (protected)
	users := api.Group("/users", auth)
	{
		users.GET("/me", userHandler.GetMe)
		users.GET("", requireAdmin, userHandler.List)
		users.PUT("/:id", requireAdmin, userHandler.Update)
		users.DELETE("/:id", requireAdmin, userHandler.Delete)
	}

	// Categories (public read, protected write)
	api.GET("/categories", categoryHandler.List)
	api.GET("/categories/:id", categoryHandler.GetByID)
	api.GET("/categories/slug/:slug", categoryHandler.GetBySlug)
	catWrite := api.Group("/categories", auth, requireEditor)
	{
		catWrite.POST("", categoryHandler.Create)
		catWrite.PUT("/:id", categoryHandler.Update)
		catWrite.DELETE("/:id", categoryHandler.Delete)
	}

	// Tags (public read, protected write)
	api.GET("/tags", tagHandler.List)
	tagWrite := api.Group("/tags", auth, requireEditor)
	{
		tagWrite.POST("", tagHandler.Create)
		tagWrite.PUT("/:id", tagHandler.Update)
		tagWrite.DELETE("/:id", tagHandler.Delete)
	}

	// Articles (public read, protected write)
	api.GET("/articles/published", articleHandler.ListPublished)
	api.GET("/articles/slug/:slug", articleHandler.GetBySlug)
	api.GET("/articles", articleHandler.List)
	api.GET("/articles/:id", articleHandler.GetByID)
	articleWrite := api.Group("/articles", auth, requireAuthor)
	{
		articleWrite.POST("", articleHandler.Create)
		articleWrite.PUT("/:id", articleHandler.Update)
		articleWrite.PATCH("/:id/status", articleHandler.UpdateStatus)
		articleWrite.DELETE("/:id", articleHandler.Delete)
	}

	// Comments (public read, protected write)
	api.GET("/articles/:id/comments", commentHandler.ListByArticle)
	commentWrite := api.Group("/comments", auth)
	{
		commentWrite.POST("", commentHandler.Create)
		commentWrite.DELETE("/:id", commentHandler.Delete)
	}

	// Uploads (protected, author+)
	uploadWrite := api.Group("/uploads", auth, requireAuthor)
	{
		uploadWrite.POST("/images", uploadHandler.UploadImage)
	}

	// Websites (public read, admin write)
	api.GET("/websites", websiteHandler.List)
	api.GET("/websites/slug/:slug", websiteHandler.GetBySlug)
	api.GET("/websites/:id", websiteHandler.GetByID)
	websiteWrite := api.Group("/websites", auth, requireAdmin)
	{
		websiteWrite.POST("", websiteHandler.Create)
		websiteWrite.PUT("/:id", websiteHandler.Update)
		websiteWrite.DELETE("/:id", websiteHandler.Delete)
	}

	return r
}
