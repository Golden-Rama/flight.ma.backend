package repository

import (
	"flight.ma.backend/src/entity"
	"gorm.io/gorm"
)

// UserRepository interface
type UserRepository interface {
	FindByID(id uint64) (*entity.User, error)
	FindByUsername(username string) (*entity.User, error)
	FindAll() ([]entity.User, error)
	Create(user *entity.User) error
	Update(user *entity.User) error
	Delete(id uint64) error
	AssignRole(userID uint64, roleID uint64) error
	ClearRoles(userID uint64) error
}

type mysqlUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &mysqlUserRepository{db: db}
}

func (r *mysqlUserRepository) FindByID(id uint64) (*entity.User, error) {
	var user entity.User
	err := r.db.Preload("Roles").First(&user, id).Error
	return &user, err
}

func (r *mysqlUserRepository) FindByUsername(username string) (*entity.User, error) {
	var user entity.User
	err := r.db.Preload("Roles").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *mysqlUserRepository) FindAll() ([]entity.User, error) {
	var users []entity.User
	err := r.db.Preload("Roles").Find(&users).Error
	return users, err
}

func (r *mysqlUserRepository) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *mysqlUserRepository) Update(user *entity.User) error {
	return r.db.Save(user).Error
}

func (r *mysqlUserRepository) Delete(id uint64) error {
	return r.db.Delete(&entity.User{}, id).Error
}

func (r *mysqlUserRepository) AssignRole(userID uint64, roleID uint64) error {
	return r.db.Exec("INSERT INTO users_role_ma (user_id, role_id) VALUES (?, ?)", userID, roleID).Error
}

func (r *mysqlUserRepository) ClearRoles(userID uint64) error {
	return r.db.Exec("DELETE FROM users_role_ma WHERE user_id = ?", userID).Error
}

// RoleRepository interface
type RoleRepository interface {
	FindAll() ([]entity.Role, error)
	FindByID(id uint64) (*entity.Role, error)
}

type mysqlRoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &mysqlRoleRepository{db: db}
}

func (r *mysqlRoleRepository) FindAll() ([]entity.Role, error) {
	var roles []entity.Role
	err := r.db.Find(&roles).Error
	return roles, err
}

func (r *mysqlRoleRepository) FindByID(id uint64) (*entity.Role, error) {
	var role entity.Role
	err := r.db.First(&role, id).Error
	return &role, err
}

// ProviderRepository interface
type ProviderRepository interface {
	FindAll(isActiveOnly bool) ([]entity.FlightProvider, error)
	FindByID(id uint64) (*entity.FlightProvider, error)
	FindByCode(code string) (*entity.FlightProvider, error)
	Create(provider *entity.FlightProvider) error
	Update(provider *entity.FlightProvider) error
	Delete(id uint64) error
}

type mysqlProviderRepository struct {
	db *gorm.DB
}

func NewProviderRepository(db *gorm.DB) ProviderRepository {
	return &mysqlProviderRepository{db: db}
}

func (r *mysqlProviderRepository) FindAll(isActiveOnly bool) ([]entity.FlightProvider, error) {
	var providers []entity.FlightProvider
	query := r.db.Preload("FlightQuest").Preload("FlightBooking")
	if isActiveOnly {
		query = query.Where("is_active = ?", true)
	}
	err := query.Find(&providers).Error
	return providers, err
}

func (r *mysqlProviderRepository) FindByID(id uint64) (*entity.FlightProvider, error) {
	var provider entity.FlightProvider
	err := r.db.Preload("FlightQuest").Preload("FlightBooking").First(&provider, id).Error
	return &provider, err
}

func (r *mysqlProviderRepository) FindByCode(code string) (*entity.FlightProvider, error) {
	var provider entity.FlightProvider
	err := r.db.Preload("FlightQuest").Preload("FlightBooking").Where("code = ?", code).First(&provider).Error
	return &provider, err
}

func (r *mysqlProviderRepository) Create(provider *entity.FlightProvider) error {
	return r.db.Create(provider).Error
}

func (r *mysqlProviderRepository) Update(provider *entity.FlightProvider) error {
	return r.db.Save(provider).Error
}

func (r *mysqlProviderRepository) Delete(id uint64) error {
	return r.db.Delete(&entity.FlightProvider{}, id).Error
}

// QuestRepository interface
type QuestRepository interface {
	FindAll() ([]entity.FlightQuest, error)
	FindByID(id uint64) (*entity.FlightQuest, error)
	Create(quest *entity.FlightQuest) error
	Update(quest *entity.FlightQuest) error
	Delete(id uint64) error
}

type mysqlQuestRepository struct {
	db *gorm.DB
}

func NewQuestRepository(db *gorm.DB) QuestRepository {
	return &mysqlQuestRepository{db: db}
}

func (r *mysqlQuestRepository) FindAll() ([]entity.FlightQuest, error) {
	var quests []entity.FlightQuest
	err := r.db.Find(&quests).Error
	return quests, err
}

func (r *mysqlQuestRepository) FindByID(id uint64) (*entity.FlightQuest, error) {
	var quest entity.FlightQuest
	err := r.db.First(&quest, id).Error
	return &quest, err
}

func (r *mysqlQuestRepository) Create(quest *entity.FlightQuest) error {
	return r.db.Create(quest).Error
}

func (r *mysqlQuestRepository) Update(quest *entity.FlightQuest) error {
	return r.db.Save(quest).Error
}

func (r *mysqlQuestRepository) Delete(id uint64) error {
	return r.db.Delete(&entity.FlightQuest{}, id).Error
}

// BookingRepository interface
type BookingRepository interface {
	FindAll() ([]entity.FlightBooking, error)
	FindByID(id uint64) (*entity.FlightBooking, error)
	Create(booking *entity.FlightBooking) error
	Update(booking *entity.FlightBooking) error
	Delete(id uint64) error
}

type mysqlBookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &mysqlBookingRepository{db: db}
}

func (r *mysqlBookingRepository) FindAll() ([]entity.FlightBooking, error) {
	var bookings []entity.FlightBooking
	err := r.db.Find(&bookings).Error
	return bookings, err
}

func (r *mysqlBookingRepository) FindByID(id uint64) (*entity.FlightBooking, error) {
	var booking entity.FlightBooking
	err := r.db.First(&booking, id).Error
	return &booking, err
}

func (r *mysqlBookingRepository) Create(booking *entity.FlightBooking) error {
	return r.db.Create(booking).Error
}

func (r *mysqlBookingRepository) Update(booking *entity.FlightBooking) error {
	return r.db.Save(booking).Error
}

func (r *mysqlBookingRepository) Delete(id uint64) error {
	return r.db.Delete(&entity.FlightBooking{}, id).Error
}
