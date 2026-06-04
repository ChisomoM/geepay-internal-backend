package db

import (
	"fmt"
	"sync"

	config "backend/pkg"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Manager is a singleton that holds the database connection pool.
// The same connection pool is shared across all requests and middleware.
// Multi-tenancy is enforced at the query level via CompanyMiddleware.
type Manager struct {
	db *gorm.DB
	mu sync.RWMutex
}

var (
	instance *Manager
	once     sync.Once
)

// Initialize creates the singleton database connection pool.
// This is called once at application startup in main().
func Initialize(cfg *config.Config, logger *zap.SugaredLogger) error {
	var err error
	once.Do(func() {
		instance, err = newManager(cfg, logger)
	})
	return err
}

// newManager creates a new database connection pool.
func newManager(cfg *config.Config, logger *zap.SugaredLogger) (*Manager, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUsername, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}

	logger.Infof("Connected to database: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)

	// Configure connection pool
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(cfg.DBPoolMaxOpen)
	sqlDB.SetMaxIdleConns(cfg.DBPoolMaxIdle)

	return &Manager{db: db}, nil
}

// GetDB returns the singleton database connection pool.
// In multi-company mode, this is scoped by CompanyMiddleware (WHERE company_id = ?).
// In single-company mode, this is used directly by handlers.
func GetDB() *gorm.DB {
	if instance == nil {
		panic("database not initialized; call db.Initialize() first")
	}
	instance.mu.RLock()
	defer instance.mu.RUnlock()
	return instance.db
}

// Close closes the database connection pool.
// Called on application shutdown.
func Close() error {
	if instance == nil {
		return nil
	}
	instance.mu.Lock()
	defer instance.mu.Unlock()
	sqlDB, err := instance.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
