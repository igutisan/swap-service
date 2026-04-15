package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"swap/iguti/swap-service/internal/config"
	"swap/iguti/swap-service/internal/controller"
	"swap/iguti/swap-service/internal/domain"
	"swap/iguti/swap-service/internal/middleware"
	"swap/iguti/swap-service/internal/migrations"
	"swap/iguti/swap-service/internal/repository"
	"swap/iguti/swap-service/internal/service"
)

func main() {
	// Configurar modo de Gin
	ginMode := getEnv("GIN_MODE", "debug")
	gin.SetMode(ginMode)

	// Conectar a la base de datos
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}

	// Migrar modelos
	if err := migrateModels(db); err != nil {
		log.Fatalf("Error al migrar modelos: %v", err)
	}

	// Ejecutar migraciones espaciales (PostGIS)
	if err := migrations.RunMigrations(db); err != nil {
		log.Printf("Advertencia: No se pudieron crear índices espaciales: %v", err)
	}

	// Configurar dependencias (Dependency Injection)
	repositories := setupRepositories(db)
	txManager := repository.NewTransactionManager(db)
	services := setupServices(repositories, txManager)
	controllers := setupControllers(services)

	// Configurar router (pasar userRepo y companyRepo para validación en middlewares)
	router := setupRouter(controllers, repositories.UserRepo, repositories.CompanyRepo)

	// Iniciar servidor
	port := getEnv("PORT", "8080")
	log.Printf("Servidor iniciado en el puerto %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

// Repositories contiene todos los repositorios de la aplicación
type Repositories struct {
	UserRepo         domain.UserRepository
	CompanyRepo      domain.CompanyRepository
	DishRepo         domain.DishRepository
	SwipeSessionRepo domain.SwipeSessionRepository
	SwipeActionRepo  domain.SwipeActionRepository
	MatchRepo        domain.MatchRepository
}

// Services contiene todos los servicios de la aplicación
type Services struct {
	AuthService         domain.AuthService
	DishService         domain.DishService
	SwipeSessionService domain.SwipeSessionService
	UploadService       service.UploadService
	GeocodingService    service.GeocodingService
}

// Controllers contiene todos los controladores de la aplicación
type Controllers struct {
	AuthController         *controller.AuthController
	DishController         *controller.DishController
	SwipeSessionController *controller.SwipeSessionController
	UploadController       *controller.UploadController
}

func setupDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "swap_service"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_SSLMODE", "disable"),
	)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	// Configurar pool de conexiones
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func migrateModels(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.User{},
		&domain.Company{},
		&domain.Dish{},
		&domain.SwipeSession{},
		&domain.SwipeAction{},
		&domain.Match{},
	)
}

func setupRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		UserRepo:         repository.NewUserRepository(db),
		CompanyRepo:      repository.NewCompanyRepository(db),
		DishRepo:         repository.NewDishRepository(db),
		SwipeSessionRepo: repository.NewSwipeSessionRepository(db),
		SwipeActionRepo:  repository.NewSwipeActionRepository(db),
		MatchRepo:        repository.NewMatchRepository(db),
	}
}

func setupServices(repos *Repositories, txManager interface{}) *Services {
	cloudinaryConfig := config.NewCloudinaryConfig()
	uploadService, err := service.NewCloudinaryService(cloudinaryConfig)
	if err != nil {
		log.Printf("Advertencia: Cloudinary no está configurado: %v", err)
		uploadService = nil
	}

	geocodingService := service.NewGeocodingService()

	return &Services{
		AuthService:         service.NewAuthService(repos.UserRepo, repos.CompanyRepo, geocodingService),
		DishService:         service.NewDishService(repos.DishRepo, repos.CompanyRepo),
		SwipeSessionService: service.NewSwipeSessionService(repos.SwipeSessionRepo, repos.DishRepo, repos.SwipeActionRepo, repos.MatchRepo, txManager),
		UploadService:       uploadService,
		GeocodingService:    geocodingService,
	}
}

func setupControllers(services *Services) *Controllers {
	uploadCtrl := controller.NewUploadController(services.UploadService)

	return &Controllers{
		AuthController:         controller.NewAuthController(services.AuthService),
		DishController:         controller.NewDishController(services.DishService),
		SwipeSessionController: controller.NewSwipeSessionController(services.SwipeSessionService),
		UploadController:       uploadCtrl,
	}
}

func setupRouter(controllers *Controllers, userRepo domain.UserRepository, companyRepo domain.CompanyRepository) *gin.Engine {
	router := gin.New()

	// Middlewares globales
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(corsMiddleware())

	// Ruta de health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Grupo de rutas API v1
	apiV1 := router.Group("/api/v1")
	{
		// Configurar middlewares de autenticación
		userAuthMiddleware := middleware.AuthMiddleware(userRepo)              // Para usuarios
		companyAuthMiddleware := middleware.CompanyAuthMiddleware(companyRepo) // Para empresas

		// Rutas de autenticación (públicas)
		controllers.AuthController.SetupRoutes(apiV1)

		// Rutas de perfil (protegidas)
		controllers.AuthController.SetupProfileRoutes(apiV1, userAuthMiddleware, companyAuthMiddleware)

		// Rutas de upload (públicas)
		controllers.UploadController.SetupRoutes(apiV1)

		// Rutas de platos (protegidas - solo empresas)
		controllers.DishController.SetupRoutes(apiV1, companyAuthMiddleware)

		// Rutas de sesiones de swipe (protegidas - solo usuarios)
		controllers.SwipeSessionController.SetupRoutes(apiV1, userAuthMiddleware)
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
