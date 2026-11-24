package server

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-todo/modular-monolith/config"
	auditApp "github.com/golang-todo/modular-monolith/internal/modules/audit/application"
	auditDomain "github.com/golang-todo/modular-monolith/internal/modules/audit/domain"
	auditInfra "github.com/golang-todo/modular-monolith/internal/modules/audit/infrastructure"
	identityApp "github.com/golang-todo/modular-monolith/internal/modules/identity/application"
	identityDomain "github.com/golang-todo/modular-monolith/internal/modules/identity/domain"
	identityInfra "github.com/golang-todo/modular-monolith/internal/modules/identity/infrastructure"
	todoApp "github.com/golang-todo/modular-monolith/internal/modules/todo/application"
	todoDomain "github.com/golang-todo/modular-monolith/internal/modules/todo/domain"
	todoInfra "github.com/golang-todo/modular-monolith/internal/modules/todo/infrastructure"
	"github.com/golang-todo/modular-monolith/internal/server/handlers"
	"github.com/golang-todo/modular-monolith/internal/server/middleware"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Server struct {
	App              *fiber.App
	DB               *gorm.DB
	Config           *config.Config
	OutboxDispatcher *auditInfra.OutboxDispatcher
}

func NewServer(cfg *config.Config) (*Server, error) {
	// Ensure SQLite storage directory exists
	dir := filepath.Dir(cfg.DatabasePath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	// Initialize GORM with SQLite
	db, err := gorm.Open(sqlite.Open(cfg.DatabasePath), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// Auto-migrate tables
	if err := db.AutoMigrate(
		&identityInfra.UserModel{},
		&auditInfra.AuditRecordModel{},
		&auditInfra.OutboxMessageModel{},
		&todoInfra.TodoModel{},
	); err != nil {
		return nil, err
	}

	// Initialize Fiber App
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.GlobalErrorHandler,
		AppName:      "Golang Todo Modular Monolith",
	})

	// Global Middlewares
	app.Use(recover.New())
	corsOrigins := cfg.CORSOrigin
	allowCredentials := false
	if corsOrigins != "*" {
		allowCredentials = true
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Trace-Id",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: allowCredentials,
	}))
	app.Use(middleware.ObservabilityMiddleware())

	// Health and Readiness
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":    "UP",
			"timestamp": time.Now().UTC(),
		})
	})
	app.Get("/ready", func(c *fiber.Ctx) error {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "DOWN", "error": "database unreachable"})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "READY"})
	})

	// Prometheus Metrics Endpoint
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// Initialize Services & Repositories
	userRepo := identityInfra.NewGormUserRepository(db)
	tokenService := identityApp.NewJWTTokenService(cfg.JWTSecret)
	identityService := identityApp.NewIdentityService(userRepo, tokenService)
	policyEngine := identityApp.NewPolicyEngine()

	auditRepo := auditInfra.NewGormAuditRepository(db)
	auditService := auditApp.NewAuditService(auditRepo)
	outboxDispatcher := auditInfra.NewOutboxDispatcher(db)

	todoRepo := todoInfra.NewGormTodoRepository(db)
	todoService := todoApp.NewTodoService(todoRepo, policyEngine, auditService, outboxDispatcher)

	// Seed Initial Data (Admin & Sample Todos if empty)
	seedDatabase(userRepo, todoRepo, auditService)

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(identityService)
	todoHandler := handlers.NewTodoHandler(todoService)
	auditHandler := handlers.NewAuditHandler(auditService)

	// API Routing
	api := app.Group("/api/v1")

	// Auth routes (Public)
	authGroup := api.Group("/auth")
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Get("/me", middleware.AuthMiddleware(tokenService), authHandler.Me)

	// Todo routes (Protected)
	todoGroup := api.Group("/todos", middleware.AuthMiddleware(tokenService))
	todoGroup.Post("/", todoHandler.Create)
	todoGroup.Get("/", todoHandler.List)
	todoGroup.Get("/stats", todoHandler.Stats)
	todoGroup.Get("/:id", todoHandler.GetByID)
	todoGroup.Put("/:id", todoHandler.Update)
	todoGroup.Patch("/:id/status", todoHandler.ChangeStatus)
	todoGroup.Delete("/:id", todoHandler.Delete)

	// Audit routes (Protected, Role-Restricted to ADMIN or AUDITOR)
	auditGroup := api.Group("/audit",
		middleware.AuthMiddleware(tokenService),
		middleware.RequireRoles(identityDomain.RoleAdmin, identityDomain.RoleAuditor),
	)
	auditGroup.Get("/", auditHandler.List)
	auditGroup.Get("/verify", auditHandler.Verify)

	return &Server{
		App:              app,
		DB:               db,
		Config:           cfg,
		OutboxDispatcher: outboxDispatcher,
	}, nil
}

func seedDatabase(userRepo identityDomain.UserRepository, todoRepo todoDomain.TodoRepository, auditService *auditApp.AuditService) {
	// Seed default Admin if not exists
	existingAdmin, _ := userRepo.FindByEmail("admin@todo.local")
	if existingAdmin == nil {
		adminUser, err := identityDomain.NewUser(
			uuid.New().String(),
			"default-tenant",
			"admin@todo.local",
			"AdminPass123!",
			"Administrator",
			identityDomain.RoleAdmin,
		)
		if err == nil {
			_ = userRepo.Save(adminUser)

			// Seed sample todos
			now := time.Now().UTC()
			tomorrow := now.Add(24 * time.Hour)
			nextWeek := now.Add(7 * 24 * time.Hour)

			todo1, _ := todoDomain.NewTodo(
				uuid.New().String(),
				"default-tenant",
				adminUser.ID,
				"Set up OpenTelemetry, Prometheus & Grafana",
				"Instrument services with traces, metrics, and dashboards.",
				todoDomain.PriorityUrgent,
				&tomorrow,
				"infrastructure,observability",
			)
			_ = todoRepo.Save(todo1)

			todo2, _ := todoDomain.NewTodo(
				uuid.New().String(),
				"default-tenant",
				adminUser.ID,
				"Implement Modular Monolith Domain Invariants",
				"Enforce state transitions and append-only tamper-evident audit logs.",
				todoDomain.PriorityHigh,
				&nextWeek,
				"backend,ddd",
			)
			_ = todoRepo.Save(todo2)

			todo3, _ := todoDomain.NewTodo(
				uuid.New().String(),
				"default-tenant",
				adminUser.ID,
				"Next.js App with Material-UI Theme",
				"Modern responsive UI with status filtering and audit log inspector.",
				todoDomain.PriorityMedium,
				&nextWeek,
				"frontend,mui",
			)
			_ = todoRepo.Save(todo3)

			// Log creation
			_ = auditService.Log(context.Background(), auditApp.RecordAuditCommand{
				TenantID:   "default-tenant",
				ActorID:    adminUser.ID,
				ActorEmail: adminUser.Email,
				Action:     auditDomain.ActionCreated,
				EntityType: "SEED",
				EntityID:   adminUser.ID,
				Data:       map[string]string{"message": "System seeded with default admin and initial tasks"},
			})
		}
	}
}
