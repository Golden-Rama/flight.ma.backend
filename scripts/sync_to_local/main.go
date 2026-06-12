package main

import (
	"fmt"
	"log"
	"os"

	"flight.ma.backend/src/entity"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Connection string for Staging MySQL (from RDS)
	mysqlDsn := "admin:ipgeIwF6OHOAZEqPF5FV@tcp(gr-services-database.cmsxjgoigtvz.ap-southeast-1.rds.amazonaws.com:3306)/gr-flight-service?charset=utf8mb4&parseTime=True&loc=Local"
	mysqlDb, err := gorm.Open(mysql.Open(mysqlDsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Staging MySQL (RDS): %v", err)
	}
	fmt.Println("Successfully connected to Staging MySQL.")

	// Get local PostgreSQL credentials from environment or use defaults
	pgUser := getEnv("PG_USER", "postgres")
	pgPass := getEnv("PG_PASSWORD", "postgres")
	pgHost := getEnv("PG_HOST", "localhost")
	pgPort := getEnv("PG_PORT", "5432")
	pgDbName := getEnv("PG_DBNAME", "gr-flight-service")

	// 2. Connect to local PostgreSQL (default db 'postgres' to check/create target database)
	pgDefaultDsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable TimeZone=Asia/Jakarta",
		pgHost, pgPort, pgUser, pgPass)
	pgDefaultDb, err := gorm.Open(postgres.Open(pgDefaultDsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to local PostgreSQL (default DB): %v", err)
	}
	fmt.Println("Successfully connected to local PostgreSQL.")

	// Create the target database if it doesn't exist
	var exists int
	pgDefaultDb.Raw("SELECT 1 FROM pg_database WHERE datname = ?", pgDbName).Scan(&exists)
	if exists != 1 {
		fmt.Printf("Creating database '%s'...\n", pgDbName)
		// CREATE DATABASE cannot run inside a transaction block, so we execute it directly
		err = pgDefaultDb.Exec(fmt.Sprintf("CREATE DATABASE \"%s\"", pgDbName)).Error
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		fmt.Printf("Database '%s' created successfully.\n", pgDbName)
	}

	// 3. Connect to the target PostgreSQL database
	pgTargetDsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
		pgHost, pgPort, pgUser, pgPass, pgDbName)
	pgDb, err := gorm.Open(postgres.Open(pgTargetDsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to target PostgreSQL database '%s': %v", pgDbName, err)
	}

	// 4. Auto Migrate target database structure
	fmt.Println("Migrating local PostgreSQL tables...")
	err = pgDb.AutoMigrate(
		&entity.User{},
		&entity.Role{},
		&entity.FlightQuest{},
		&entity.FlightBooking{},
		&entity.FlightProvider{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	fmt.Println("Migration completed successfully.")

	// 5. Sync Roles
	fmt.Println("Syncing roles...")
	var roles []entity.Role
	if err := mysqlDb.Find(&roles).Error; err == nil {
		for _, r := range roles {
			pgDb.Save(&r)
		}
	}

	// 6. Sync Flight Quests
	fmt.Println("Syncing flight quests...")
	var quests []entity.FlightQuest
	if err := mysqlDb.Find(&quests).Error; err == nil {
		for _, q := range quests {
			pgDb.Save(&q)
		}
	}

	// 7. Sync Flight Bookings
	fmt.Println("Syncing flight bookings...")
	var bookings []entity.FlightBooking
	if err := mysqlDb.Find(&bookings).Error; err == nil {
		for _, b := range bookings {
			pgDb.Save(&b)
		}
	}

	// 8. Sync Flight Providers
	fmt.Println("Syncing flight providers...")
	var providers []entity.FlightProvider
	if err := mysqlDb.Find(&providers).Error; err == nil {
		for _, p := range providers {
			pgDb.Save(&p)
		}
	}

	// 9. Sync Users
	fmt.Println("Syncing users...")
	var users []entity.User
	if err := mysqlDb.Find(&users).Error; err == nil {
		for _, u := range users {
			pgDb.Save(&u)
		}
	}
	
	// Sync many-to-many relationship users_role_ma
	type UserRoleMa struct {
		UserID uint64 `gorm:"column:user_id"`
		RoleID uint64 `gorm:"column:role_id"`
	}
	var userRoles []UserRoleMa
	if err := mysqlDb.Table("users_role_ma").Find(&userRoles).Error; err == nil {
		fmt.Println("Syncing user roles relations...")
		for _, ur := range userRoles {
			pgDb.Exec("INSERT INTO users_role_ma (user_id, role_id) VALUES (?, ?) ON CONFLICT DO NOTHING", ur.UserID, ur.RoleID)
		}
	}

	fmt.Println("Data sync completed successfully!")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
