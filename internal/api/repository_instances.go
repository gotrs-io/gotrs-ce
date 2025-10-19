package api

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gotrs-io/gotrs-ce/internal/config"
	"github.com/gotrs-io/gotrs-ce/internal/database"
	"github.com/gotrs-io/gotrs-ce/internal/repository"
	"github.com/gotrs-io/gotrs-ce/internal/service"
	"github.com/gotrs-io/gotrs-ce/internal/shared"
)

var (
	ticketRepo          repository.ITicketRepository
	queueRepo           *repository.QueueRepository
	priorityRepo        *repository.PriorityRepository
	userRepo            *repository.UserRepository
	simpleTicketService *service.SimpleTicketService
	storageService      service.StorageService
	lookupService       *service.LookupService
	authService         *service.AuthService
	once                sync.Once
)

// InitializeServices initializes singleton service instances
func InitializeServices() {
	once.Do(func() {
		// In test without DB config, initialize lightweight services without DB
		if os.Getenv("APP_ENV") == "test" && os.Getenv("DB_HOST") == "" && os.Getenv("DATABASE_URL") == "" {
			// Minimal storage service
			storagePath := os.Getenv("STORAGE_PATH")
			if storagePath == "" {
				storagePath = "/tmp"
			}
			if ss, err := service.NewLocalStorageService(storagePath); err == nil {
				storageService = ss
			} else {
				log.Printf("WARNING: storage init failed in test: %v", err)
			}
			// Minimal in-memory ticket service with nil repo; guarded usage
			simpleTicketService = service.NewSimpleTicketService(nil)
			// Lookup service will self-guard on nil DB
			lookupService = service.NewLookupService()
			return
		}

		// Get database connection - no fallback in prod; in tests allow DB-less mode
		db, err := database.GetDB()
		if err != nil || db == nil {
			if os.Getenv("APP_ENV") == "test" {
				// Fallback to lightweight services
				storagePath := os.Getenv("STORAGE_PATH")
				if storagePath == "" {
					storagePath = "/tmp"
				}
				if ss, e := service.NewLocalStorageService(storagePath); e == nil {
					storageService = ss
				} else {
					log.Printf("WARNING: storage init failed in test: %v", e)
				}
				simpleTicketService = service.NewSimpleTicketService(nil)
				lookupService = service.NewLookupService()
				return
			}
			log.Fatalf("FATAL: Cannot initialize services without database connection: %v", err)
		}

		log.Printf("Successfully connected to database")
		// Initialize real database repositories
		ticketRepo = repository.NewTicketRepository(db)
		queueRepo = repository.NewQueueRepository(db)
		priorityRepo = repository.NewPriorityRepository(db)
		userRepo = repository.NewUserRepository(db)

		// Initialize services
		simpleTicketService = service.NewSimpleTicketService(ticketRepo)

		// Initialize lookup service
		lookupService = service.NewLookupService()

		// Initialize storage service from config
		cfg := config.Get()
		if cfg != nil && cfg.Storage.Type == "db" {
			storageService, err = service.NewDatabaseStorageService()
			if err != nil {
				log.Fatalf("FATAL: Cannot initialize DB storage service: %v", err)
			}
			log.Printf("StorageService: using DB backend")
		} else {
			storagePath := resolveStoragePath(cfg)
			storageService, err = service.NewLocalStorageService(storagePath)
			if err != nil {
				log.Fatalf("FATAL: Cannot initialize storage service: %v", err)
			}
			log.Printf("StorageService: using local backend at %s", storagePath)
		}

		// Initialize auth service
		// Use the shared JWT manager for auth service
		jwtManager := shared.GetJWTManager()
		authService = service.NewAuthService(db, jwtManager)
	})
}

// GetTicketService returns the singleton simple ticket service instance
func GetTicketService() *service.SimpleTicketService {
	InitializeServices()
	return simpleTicketService
}

// GetStorageService returns the singleton storage service instance
func GetStorageService() service.StorageService {
	InitializeServices()
	return storageService
}

// GetTicketRepository returns the singleton ticket repository instance
func GetTicketRepository() repository.ITicketRepository {
	InitializeServices()
	return ticketRepo
}

// GetLookupService returns the singleton lookup service instance
func GetLookupService() *service.LookupService {
	InitializeServices()
	return lookupService
}

// GetQueueRepository returns the singleton queue repository instance
func GetQueueRepository() *repository.QueueRepository {
	InitializeServices()
	return queueRepo
}

// GetPriorityRepository returns the singleton priority repository instance
func GetPriorityRepository() *repository.PriorityRepository {
	InitializeServices()
	return priorityRepo
}

// GetUserRepository returns the singleton user repository instance
func GetUserRepository() *repository.UserRepository {
	InitializeServices()
	return userRepo
}

// GetAuthService returns the singleton auth service instance
func GetAuthService() *service.AuthService {
	InitializeServices()
	return authService
}

func resolveStoragePath(cfg *config.Config) string {
	if env := os.Getenv("STORAGE_PATH"); env != "" {
		return env
	}

	if cfg != nil && cfg.Storage.Local.Path != "" {
		return cfg.Storage.Local.Path
	}

	if isTestProcess() {
		if tmp, err := os.MkdirTemp("", "gotrs-storage-"); err == nil {
			return tmp
		}
	}

	if root := findRepoRoot(); root != "" {
		return filepath.Join(root, "storage")
	}

	return "./storage"
}

func isTestProcess() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}

func findRepoRoot() string {
	dir, err := os.Getwd()
	for err == nil {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}
