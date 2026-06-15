package main

import (
	"fmt"
	"log"
	"net/http"

	"flight.ma.backend/config"
	"flight.ma.backend/src/factory"
	customAuth "flight.ma.backend/src/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// 1. Load configuration
	cfg := config.Get()

	// 2. Initialize Resolver Factory (Database, Repositories, Services, Handlers)
	res, err := factory.NewResolver(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize resolver: %v", err)
	}

	// 3. Initialize Auth Middlewares
	apiAuthMiddleware := customAuth.NewApiAuthMiddleware(cfg.JwtSecret)
	dashboardAuthMiddleware := customAuth.NewDashboardAuthMiddleware()

	// 4. Initialize Unified API/Backend Server (Echo)
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())

	// CORS Middleware configured to allow requests from the separated Next.js/UmiJS frontend
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://localhost:8000", "http://127.0.0.1:8000"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Provider"},
		AllowCredentials: true,
	}))

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "API: method=${method}, uri=${uri}, status=${status}\n",
	}))

	// Health Check
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{
			"status":   true,
			"app_name": cfg.AppName + " (API Backend)",
			"version":  "1.0.0",
		})
	})

	// Register Flight API Routes (/api/v1)
	res.ApiHandler.RegisterRoutes(e, apiAuthMiddleware)

	// Register Dashboard REST API Routes (/api/dashboard)
	res.DashboardHandler.RegisterRoutes(e, dashboardAuthMiddleware)

	// Start Backend API Server
	log.Printf("Starting API Backend Server on port %d...", cfg.ApiPort)
	if err := e.Start(fmt.Sprintf(":%d", cfg.ApiPort)); err != nil {
		log.Fatalf("API Server startup failed: %v", err)
	}
}
