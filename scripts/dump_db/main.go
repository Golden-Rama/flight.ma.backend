package main

import (
	"fmt"
	"log"
	"os"

	"flight.ma.backend/config"
	"flight.ma.backend/src/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Get()

	// Connect to RDS
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to RDS: %v", err)
	}

	f, err := os.Create("init.sql")
	if err != nil {
		log.Fatalf("Failed to create init.sql: %v", err)
	}
	defer f.Close()

	// Header
	f.WriteString("CREATE DATABASE IF NOT EXISTS `gr-flight-service`;\n")
	f.WriteString("USE `gr-flight-service`;\n\n")

	// Dump flight_quests
	var quests []entity.FlightQuest
	if err := db.Find(&quests).Error; err == nil {
		f.WriteString("-- Dumping flight_quests\n")
		for _, q := range quests {
			f.WriteString(fmt.Sprintf(
				"INSERT INTO `flight_quests` (`id`, `endpoint`, `preferred_carriers_gds`, `preferred_carriers_non_gds`, `created_at`, `updated_at`) VALUES (%d, '%s', '%s', '%s', '%s', '%s') ON DUPLICATE KEY UPDATE `endpoint`=VALUES(`endpoint`);\n",
				q.ID, q.Endpoint, q.PreferredCarriersGds, q.PreferredCarriersNonGds, q.CreatedAt.Format("2006-01-02 15:04:05"), q.UpdatedAt.Format("2006-01-02 15:04:05"),
			))
		}
		f.WriteString("\n")
	}

	// Dump flight_bookings
	var bookings []entity.FlightBooking
	if err := db.Find(&bookings).Error; err == nil {
		f.WriteString("-- Dumping flight_bookings\n")
		for _, b := range bookings {
			f.WriteString(fmt.Sprintf(
				"INSERT INTO `flight_bookings` (`id`, `endpoint_fare_detail`, `endpoint_booking`, `endpoint_check_booking`, `endpoint_issue_ticket`, `endpoint_cancel_booking`, `created_at`, `updated_at`) VALUES (%d, '%s', '%s', '%s', '%s', '%s', '%s', '%s') ON DUPLICATE KEY UPDATE `endpoint_booking`=VALUES(`endpoint_booking`);\n",
				b.ID, b.EndpointFareDetail, b.EndpointBooking, b.EndpointCheckBooking, b.EndpointIssueTicket, b.EndpointCancelBooking, b.CreatedAt.Format("2006-01-02 15:04:05"), b.UpdatedAt.Format("2006-01-02 15:04:05"),
			))
		}
		f.WriteString("\n")
	}

	// Dump flight_providers
	var providers []entity.FlightProvider
	if err := db.Find(&providers).Error; err == nil {
		f.WriteString("-- Dumping flight_providers\n")
		for _, p := range providers {
			questIdStr := "NULL"
			if p.FlightQuestId != nil {
				questIdStr = fmt.Sprintf("%d", *p.FlightQuestId)
			}
			bookingIdStr := "NULL"
			if p.FlightReservationId != nil {
				bookingIdStr = fmt.Sprintf("%d", *p.FlightReservationId)
			}
			activeVal := 0
			if p.IsActive {
				activeVal = 1
			}
			domesticVal := 0
			if p.AvailableDomestic {
				domesticVal = 1
			}
			f.WriteString(fmt.Sprintf(
				"INSERT INTO `flight_providers` (`id`, `name`, `code`, `base_url`, `is_active`, `available_domestic`, `description`, `created_at`, `updated_at`, `flight_quest_id`, `flight_reservation_id`) VALUES (%d, '%s', '%s', '%s', %d, %d, '%s', '%s', '%s', %s, %s) ON DUPLICATE KEY UPDATE `name`=VALUES(`name`);\n",
				p.ID, p.Name, p.Code, p.BaseUrl, activeVal, domesticVal, p.Description, p.CreatedAt.Format("2006-01-02 15:04:05"), p.UpdatedAt.Format("2006-01-02 15:04:05"), questIdStr, bookingIdStr,
			))
		}
		f.WriteString("\n")
	}

	fmt.Println("Database dump completed successfully to init.sql!")
}
