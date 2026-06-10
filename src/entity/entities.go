package entity

import (
	"time"
)

type User struct {
	ID        uint64    `gorm:"primaryKey;column:id" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	Username  string    `gorm:"column:username;unique" json:"username"`
	Password  string    `gorm:"column:password" json:"password"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	Roles     []Role    `gorm:"many2many:users_role_ma;foreignKey:id;joinForeignKey:user_id;references:id;joinReferences:role_id" json:"roles"`
}

func (User) TableName() string {
	return "users_ma"
}

type Role struct {
	ID          uint64 `gorm:"primaryKey;column:id" json:"id"`
	Name        string `gorm:"column:name" json:"name"`
	Code        string `gorm:"column:code;unique" json:"code"`
	Description string `gorm:"column:description" json:"description"`
}

func (Role) TableName() string {
	return "roles_ma"
}

type FlightQuest struct {
	ID                      uint64    `gorm:"primaryKey;column:id" json:"id"`
	Endpoint                string    `gorm:"column:endpoint" json:"endpoint"`
	PreferredCarriersGds    string    `gorm:"column:preferred_carriers_gds" json:"preferred_carriers_gds"`
	PreferredCarriersNonGds string    `gorm:"column:preferred_carriers_non_gds" json:"preferred_carriers_non_gds"`
	CreatedAt               time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt               time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (FlightQuest) TableName() string {
	return "flight_quests"
}

type FlightBooking struct {
	ID                    uint64    `gorm:"primaryKey;column:id" json:"id"`
	EndpointFareDetail    string    `gorm:"column:endpoint_fare_detail" json:"endpoint_fare_detail"`
	EndpointBooking       string    `gorm:"column:endpoint_booking" json:"endpoint_booking"`
	EndpointCheckBooking  string    `gorm:"column:endpoint_check_booking" json:"endpoint_check_booking"`
	EndpointIssueTicket   string    `gorm:"column:endpoint_issue_ticket" json:"endpoint_issue_ticket"`
	EndpointCancelBooking string    `gorm:"column:endpoint_cancel_booking" json:"endpoint_cancel_booking"`
	CreatedAt             time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (FlightBooking) TableName() string {
	return "flight_bookings"
}

type FlightProvider struct {
	ID                  uint64         `gorm:"primaryKey;column:id" json:"id"`
	Name                string         `gorm:"column:name" json:"name"`
	Code                string         `gorm:"column:code;unique" json:"code"`
	BaseUrl             string         `gorm:"column:base_url" json:"base_url"`
	IsActive            bool           `gorm:"column:is_active;default:1" json:"is_active"`
	AvailableDomestic   bool           `gorm:"column:available_domestic;default:1" json:"available_domestic"`
	Description         string         `gorm:"column:description" json:"description"`
	CreatedAt           time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"column:updated_at" json:"updated_at"`
	FlightQuestId       *uint64        `gorm:"column:flight_quest_id" json:"flight_quest_id"`
	FlightQuest         *FlightQuest   `gorm:"foreignKey:FlightQuestId;references:ID" json:"flight_quest,omitempty"`
	FlightReservationId *uint64        `gorm:"column:flight_reservation_id" json:"flight_reservation_id"`
	FlightBooking       *FlightBooking `gorm:"foreignKey:FlightReservationId;references:ID" json:"flight_booking,omitempty"`
}

func (FlightProvider) TableName() string {
	return "flight_providers"
}
