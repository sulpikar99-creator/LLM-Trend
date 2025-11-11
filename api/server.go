package api

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sulpikar99-creator/LLM-Trend/bootstrap"
)

// Server represents the HTTP server
type Server struct {
	Router *gin.Engine
	app    *bootstrap.Application
}

// NewServer creates a new HTTP server
func NewServer(app *bootstrap.Application) *Server {
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	server := &Server{
		Router: router,
		app:    app,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check
	s.Router.GET("/api/health", s.handleHealth)

	// Public routes
	public := s.Router.Group("/api")
	{
		public.POST("/auth/register", s.handleRegister)
		public.POST("/auth/login", s.handleLogin)
	}

	// Protected routes (require authentication)
	protected := s.Router.Group("/api")
	protected.Use(s.authMiddleware())
	{
		// User routes
		protected.GET("/user/profile", s.handleGetProfile)
		protected.PUT("/user/profile", s.handleUpdateProfile)

		// Trader routes
		protected.GET("/traders", s.handleListTraders)
		protected.POST("/traders", s.handleCreateTrader)
		protected.GET("/traders/:id", s.handleGetTrader)
		protected.PUT("/traders/:id", s.handleUpdateTrader)
		protected.DELETE("/traders/:id", s.handleDeleteTrader)
		protected.POST("/traders/:id/start", s.handleStartTrader)
		protected.POST("/traders/:id/stop", s.handleStopTrader)

		// Analytics routes
		protected.GET("/analytics/drawdown", s.handleGetDrawdown)
		protected.GET("/analytics/montecarlo", s.handleGetMonteCarlo)
		protected.GET("/analytics/correlation", s.handleGetCorrelation)
		protected.GET("/analytics/performance", s.handleGetPerformance)

		// Decision log routes
		protected.GET("/decisions/:trader_id", s.handleGetDecisions)
		protected.GET("/decisions/:trader_id/:cycle", s.handleGetDecisionByCycle)
		protected.DELETE("/decisions/:trader_id/old", s.handleDeleteOldDecisions)

		// Risk management routes
		protected.GET("/risk/:trader_id/config", s.handleGetRiskConfig)
		protected.PUT("/risk/:trader_id/config", s.handleUpdateRiskConfig)
		protected.GET("/risk/:trader_id/status", s.handleGetRiskStatus)
		protected.POST("/risk/:trader_id/reset", s.handleResetRiskMonitor)
		protected.POST("/risk/:trader_id/calculate-size", s.handleCalculatePositionSize)
		protected.POST("/risk/:trader_id/calculate-sltp", s.handleCalculateStopLossTakeProfit)

		// Config routes
		protected.GET("/config", s.handleGetConfig)
		protected.PUT("/config", s.handleUpdateConfig)

		// User config routes
		protected.GET("/user/config", s.handleGetUserConfig)
		protected.PUT("/user/config", s.handleUpdateUserConfig)
		protected.GET("/user/config/ai", s.handleGetUserAIConfig)
		protected.PUT("/user/config/ai", s.handleUpdateUserAIConfig)
		protected.GET("/user/config/prompts", s.handleGetUserPrompts)
		protected.PUT("/user/config/prompts", s.handleUpdateUserPrompts)
		protected.PUT("/user/config/exchange", s.handleUpdateExchangeConfig)
		protected.PUT("/user/config/risk-limits", s.handleUpdateRiskLimits)
		protected.PUT("/user/config/notifications", s.handleUpdateNotificationSettings)
		protected.POST("/user/config/reset", s.handleResetUserConfig)
		protected.GET("/user/ai-providers", s.handleGetAvailableAIProviders)
	}

	// Admin routes (require authentication + admin role)
	admin := s.Router.Group("/api/admin")
	admin.Use(s.authMiddleware())
	{
		// Beta code management
		admin.POST("/beta-codes", s.handleCreateBetaCode)
		admin.GET("/beta-codes", s.handleListBetaCodes)
		admin.GET("/beta-codes/stats", s.handleGetBetaCodeStats)
		admin.GET("/beta-codes/:code", s.handleGetBetaCode)
		admin.POST("/beta-codes/:code/deactivate", s.handleDeactivateBetaCode)
		admin.POST("/beta-codes/:code/reactivate", s.handleReactivateBetaCode)
		admin.DELETE("/beta-codes/:code", s.handleDeleteBetaCode)

		// User management
		admin.GET("/users", s.handleListUsers)
		admin.PUT("/users/:user_id/role", s.handleUpdateUserRole)
		admin.POST("/users/:user_id/deactivate", s.handleDeactivateUser)
		admin.POST("/users/:user_id/reactivate", s.handleReactivateUser)
	}
}

// handleHealth handles health check requests
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
	})
}

// Response helpers
func successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func errorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
	})
}
