package bootstrap

import (
	"fmt"
	"log"
	"os"

	"github.com/sulpikar99-creator/LLM-Trend/config"
	"github.com/sulpikar99-creator/LLM-Trend/crypto"
	"github.com/sulpikar99-creator/LLM-Trend/logger"
	"github.com/sulpikar99-creator/LLM-Trend/manager"
)

// Application holds all initialized components
type Application struct {
	Config        *config.Config
	Database      *config.Database
	Logger        *logger.Logger
	CryptoService *crypto.Service
	TraderManager *manager.TraderManager
}

// Initialize sets up the application and all its components
func Initialize() (*Application, error) {
	log.Println("Initializing application...")

	// Initialize logger
	appLogger := logger.New()

	// Validate environment variables
	if err := validateEnvironment(); err != nil {
		return nil, fmt.Errorf("environment validation failed: %w", err)
	}

	// Initialize crypto service
	cryptoService, err := crypto.NewService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize crypto service: %w", err)
	}

	// Initialize database
	db, err := config.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize beta code manager
	cfg.BetaCodeManager = config.NewBetaCodeManager(db.DB)

	// Initialize user config manager
	cfg.UserConfigManager = config.NewUserConfigManager(db.DB)

	// Initialize trader manager
	traderMgr := manager.NewTraderManager()

	// Set database connection for trader statistics
	traderMgr.SetDatabase(db.DB)

	// Load existing traders from database (critical for reconnecting after restart)
	if err := traderMgr.LoadTradersFromDatabase(db.DB, cryptoService); err != nil {
		log.Printf("Warning: Failed to load traders from database: %v", err)
		// Don't fail startup - we can still create new traders
	}

	app := &Application{
		Config:        cfg,
		Database:      db,
		Logger:        appLogger,
		CryptoService: cryptoService,
		TraderManager: traderMgr,
	}

	log.Println("Application initialized successfully")
	return app, nil
}

// Cleanup performs cleanup operations
func (app *Application) Cleanup() {
	log.Println("Cleaning up application...")
	if app.Database != nil {
		app.Database.Close()
	}
	log.Println("Application cleanup completed")
}

// validateEnvironment checks that required environment variables are set
func validateEnvironment() error {
	required := []string{
		"DATA_ENCRYPTION_KEY",
		"JWT_SECRET",
	}

	for _, key := range required {
		if os.Getenv(key) == "" {
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	return nil
}
