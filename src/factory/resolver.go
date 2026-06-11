package factory

import (
	"fmt"
	"time"

	"gr-flight-ma-new/config"
	"gr-flight-ma-new/src/entity"
	"gr-flight-ma-new/src/handler"
	"gr-flight-ma-new/src/repository"
	"gr-flight-ma-new/src/service"
	"gr-flight-ma-new/src/utils"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Resolver struct {
	Cfg              *config.Config
	DB               *gorm.DB
	UserRepo         repository.UserRepository
	RoleRepo         repository.RoleRepository
	ProviderRepo     repository.ProviderRepository
	QuestRepo        repository.QuestRepository
	BookingRepo      repository.BookingRepository
	SearchService    service.SearchService
	BookingService   service.BookingService
	ApiHandler       *handler.ApiHandler
	DashboardHandler *handler.DashboardHandler
}

func NewResolver(cfg *config.Config) (*Resolver, error) {
	var dialector gorm.Dialector

	if cfg.DBDriver == "postgres" || cfg.DBDriver == "psql" {
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
		dialector = postgres.Open(dsn)
	} else {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
		dialector = mysql.Open(dsn)
	}

	gormConfig := &gorm.Config{}
	if cfg.AppDebug {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(20)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
	}

	// Auto Migration to ensure database tables exist
	err = db.AutoMigrate(
		&entity.User{},
		&entity.Role{},
		&entity.FlightQuest{},
		&entity.FlightBooking{},
		&entity.FlightProvider{},
	)
	if err != nil {
		return nil, fmt.Errorf("auto migration failed: %w", err)
	}

	// Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	providerRepo := repository.NewProviderRepository(db)
	questRepo := repository.NewQuestRepository(db)
	bookingRepo := repository.NewBookingRepository(db)

	// Seed Roles and Admin User if they do not exist
	seedDatabase(db, userRepo)

	// Initialize Services
	searchService := service.NewSearchService(providerRepo)
	bookingService := service.NewBookingService(providerRepo)

	// Initialize Handlers
	apiHandler := handler.NewApiHandler(searchService, bookingService)
	dashboardHandler := handler.NewDashboardHandler(userRepo, roleRepo, providerRepo, questRepo, bookingRepo)

	return &Resolver{
		Cfg:              cfg,
		DB:               db,
		UserRepo:         userRepo,
		RoleRepo:         roleRepo,
		ProviderRepo:     providerRepo,
		QuestRepo:        questRepo,
		BookingRepo:      bookingRepo,
		SearchService:    searchService,
		BookingService:   bookingService,
		ApiHandler:       apiHandler,
		DashboardHandler: dashboardHandler,
	}, nil
}

func seedDatabase(db *gorm.DB, userRepo repository.UserRepository) {
	// Seed roles
	var count int64
	db.Model(&entity.Role{}).Count(&count)
	if count == 0 {
		roles := []entity.Role{
			{Name: "Administrator", Code: "admin", Description: "Administrator Role"},
			{Name: "Super User", Code: "su", Description: "Super User Role"},
		}
		for _, r := range roles {
			db.Create(&r)
		}
	}

	// Seed default admin user
	db.Model(&entity.User{}).Count(&count)
	if count == 0 {
		hashedPass, _ := utils.HashPassword("admin123")
		adminUser := &entity.User{
			Name:      "System Admin",
			Username:  "admin",
			Password:  hashedPass,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = userRepo.Create(adminUser)

		// Get role ID for 'su'
		var suRole entity.Role
		if err := db.Where("code = ?", "su").First(&suRole).Error; err == nil {
			_ = userRepo.AssignRole(adminUser.ID, suRole.ID)
		}
	}

	// Seed default Flight Service provider (flight.server)
	var providerCount int64
	db.Model(&entity.FlightProvider{}).Count(&providerCount)
	if providerCount == 0 {
		quest := entity.FlightQuest{
			Endpoint:                "/api/v1/search",
			PreferredCarriersGds:    "GA,SQ",
			PreferredCarriersNonGds: "JT,QG",
			CreatedAt:               time.Now(),
			UpdatedAt:               time.Now(),
		}
		db.Create(&quest)

		booking := entity.FlightBooking{
			EndpointFareDetail:    "/api/v1/fare-detail",
			EndpointBooking:       "/api/v1/reservation",
			EndpointCheckBooking:  "/api/v1/check-reservation",
			EndpointIssueTicket:   "/api/v1/issue-ticket",
			EndpointCancelBooking: "/api/v1/cancel-reservation",
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}
		db.Create(&booking)

		provider := entity.FlightProvider{
			Name:                "Flight Service",
			Code:                "flight-service",
			BaseUrl:             "http://host.docker.internal:8000",
			IsActive:            true,
			AvailableDomestic:   true,

			Description:         "Main flight service GDS & LCC provider (flight.service)",
			FlightQuestId:       &quest.ID,
			FlightReservationId: &booking.ID,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}
		db.Create(&provider)
	}
}
