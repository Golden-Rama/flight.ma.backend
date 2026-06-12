package handler

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"flight.ma.backend/src/entity"
	"flight.ma.backend/src/repository"
	"flight.ma.backend/src/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	sessionMap = make(map[string]*entity.User)
	sessionMu  sync.RWMutex
)

type DashboardHandler struct {
	userRepo     repository.UserRepository
	roleRepo     repository.RoleRepository
	providerRepo repository.ProviderRepository
	questRepo    repository.QuestRepository
	bookingRepo  repository.BookingRepository
}

func NewDashboardHandler(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	providerRepo repository.ProviderRepository,
	questRepo repository.QuestRepository,
	bookingRepo repository.BookingRepository,
) *DashboardHandler {
	return &DashboardHandler{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		providerRepo: providerRepo,
		questRepo:    questRepo,
		bookingRepo:  bookingRepo,
	}
}

func (h *DashboardHandler) RegisterRoutes(e *echo.Echo, authMiddleware echo.MiddlewareFunc) {
	// Public auth route
	e.POST("/api/dashboard/login", h.Login)
	e.POST("/api/dashboard/logout", h.Logout)

	// Protected dashboard API group
	g := e.Group("/api/dashboard", authMiddleware)
	g.GET("/me", h.Me)

	// Providers CRUD
	g.GET("/providers", h.GetProviders)
	g.POST("/providers", h.CreateProvider)
	g.GET("/providers/:id", h.GetProvider)
	g.PUT("/providers/:id", h.UpdateProvider)
	g.DELETE("/providers/:id", h.DeleteProvider)

	// Quests CRUD
	g.GET("/quests", h.GetQuests)
	g.POST("/quests", h.CreateQuest)
	g.GET("/quests/:id", h.GetQuest)
	g.PUT("/quests/:id", h.UpdateQuest)
	g.DELETE("/quests/:id", h.DeleteQuest)

	// Bookings CRUD
	g.GET("/bookings", h.GetBookings)
	g.POST("/bookings", h.CreateBooking)
	g.GET("/bookings/:id", h.GetBooking)
	g.PUT("/bookings/:id", h.UpdateBooking)
	g.DELETE("/bookings/:id", h.DeleteBooking)

	// Users CRUD
	g.GET("/users", h.GetUsers)
	g.POST("/users", h.CreateUser)
	g.GET("/users/:id", h.GetUser)
	g.PUT("/users/:id", h.UpdateUser)
	g.DELETE("/users/:id", h.DeleteUser)

	// Roles list
	g.GET("/roles", h.GetRoles)
}

func GetSessionUser(token string) *entity.User {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	return sessionMap[token]
}

func (h *DashboardHandler) Login(c echo.Context) error {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	user, err := h.userRepo.FindByUsername(input.Username)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
	}

	if !utils.VerifyPassword(input.Password, user.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
	}

	// Create session token
	token := uuid.NewString()

	sessionMu.Lock()
	sessionMap[token] = user
	sessionMu.Unlock()

	cookie := &http.Cookie{
		Name:     "admin_auth",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(2 * time.Hour),
		HttpOnly: true,
	}
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Login successful",
		"user":    user,
	})
}

func (h *DashboardHandler) Logout(c echo.Context) error {
	cookie, err := c.Cookie("admin_auth")
	if err == nil {
		sessionMu.Lock()
		delete(sessionMap, cookie.Value)
		sessionMu.Unlock()
	}

	clearCookie := &http.Cookie{
		Name:     "admin_auth",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	}
	c.SetCookie(clearCookie)

	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func (h *DashboardHandler) Me(c echo.Context) error {
	cookie, err := c.Cookie("admin_auth")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	user := GetSessionUser(cookie.Value)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	// Refetch from database to get latest roles/info
	latestUser, err := h.userRepo.FindByID(user.ID)
	if err == nil {
		return c.JSON(http.StatusOK, latestUser)
	}

	return c.JSON(http.StatusOK, user)
}

// Providers CRUD implementations
func (h *DashboardHandler) GetProviders(c echo.Context) error {
	providers, err := h.providerRepo.FindAll(false)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, providers)
}

func (h *DashboardHandler) CreateProvider(c echo.Context) error {
	var input entity.FlightProvider
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	input.Code = utils.Slugify(input.Name)
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	if err := h.providerRepo.Create(&input); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, input)
}

func (h *DashboardHandler) GetProvider(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	provider, err := h.providerRepo.FindByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Provider not found"})
	}

	return c.JSON(http.StatusOK, provider)
}

func (h *DashboardHandler) UpdateProvider(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	existing, err := h.providerRepo.FindByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Provider not found"})
	}

	var input struct {
		Name                string  `json:"name"`
		BaseUrl             string  `json:"base_url"`
		IsActive            bool    `json:"is_active"`
		AvailableDomestic   bool    `json:"available_domestic"`
		Description         string  `json:"description"`
		FlightQuestId       *uint64 `json:"flight_quest_id"`
		FlightReservationId *uint64 `json:"flight_reservation_id"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	existing.Name = input.Name
	existing.Code = utils.Slugify(input.Name)
	existing.BaseUrl = input.BaseUrl
	existing.IsActive = input.IsActive
	existing.AvailableDomestic = input.AvailableDomestic
	existing.Description = input.Description
	existing.FlightQuestId = input.FlightQuestId
	existing.FlightReservationId = input.FlightReservationId
	existing.UpdatedAt = time.Now()

	if err := h.providerRepo.Update(existing); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, existing)
}

func (h *DashboardHandler) DeleteProvider(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	if err := h.providerRepo.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Provider deleted successfully"})
}

// Quests CRUD
func (h *DashboardHandler) GetQuests(c echo.Context) error {
	quests, err := h.questRepo.FindAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, quests)
}

func (h *DashboardHandler) CreateQuest(c echo.Context) error {
	var input entity.FlightQuest
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	if err := h.questRepo.Create(&input); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, input)
}

func (h *DashboardHandler) GetQuest(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	quest, err := h.questRepo.FindByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Quest not found"})
	}

	return c.JSON(http.StatusOK, quest)
}

func (h *DashboardHandler) UpdateQuest(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	existing, err := h.questRepo.FindByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Quest not found"})
	}

	var input struct {
		Endpoint                string `json:"endpoint"`
		PreferredCarriersGds    string `json:"preferred_carriers_gds"`
		PreferredCarriersNonGds string `json:"preferred_carriers_non_gds"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	existing.Endpoint = input.Endpoint
	existing.PreferredCarriersGds = input.PreferredCarriersGds
	existing.PreferredCarriersNonGds = input.PreferredCarriersNonGds
	existing.UpdatedAt = time.Now()

	if err := h.questRepo.Update(existing); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, existing)
}

func (h *DashboardHandler) DeleteQuest(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	if err := h.questRepo.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Quest deleted successfully"})
}

// Bookings CRUD
func (h *DashboardHandler) GetBookings(c echo.Context) error {
	bookings, err := h.bookingRepo.FindAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, bookings)
}

func (h *DashboardHandler) CreateBooking(c echo.Context) error {
	var input entity.FlightBooking
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	if err := h.bookingRepo.Create(&input); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, input)
}

func (h *DashboardHandler) GetBooking(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	booking, err := h.bookingRepo.FindByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Booking configuration not found"})
	}

	return c.JSON(http.StatusOK, booking)
}

func (h *DashboardHandler) UpdateBooking(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	existing, err := h.bookingRepo.FindByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Booking configuration not found"})
	}

	var input struct {
		EndpointFareDetail    string `json:"endpoint_fare_detail"`
		EndpointBooking       string `json:"endpoint_booking"`
		EndpointCheckBooking  string `json:"endpoint_check_booking"`
		EndpointIssueTicket   string `json:"endpoint_issue_ticket"`
		EndpointCancelBooking string `json:"endpoint_cancel_booking"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	existing.EndpointFareDetail = input.EndpointFareDetail
	existing.EndpointBooking = input.EndpointBooking
	existing.EndpointCheckBooking = input.EndpointCheckBooking
	existing.EndpointIssueTicket = input.EndpointIssueTicket
	existing.EndpointCancelBooking = input.EndpointCancelBooking
	existing.UpdatedAt = time.Now()

	if err := h.bookingRepo.Update(existing); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, existing)
}

func (h *DashboardHandler) DeleteBooking(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	if err := h.bookingRepo.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Booking configuration deleted successfully"})
}

// Users CRUD
func (h *DashboardHandler) GetUsers(c echo.Context) error {
	users, err := h.userRepo.FindAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, users)
}

func (h *DashboardHandler) CreateUser(c echo.Context) error {
	var input struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Password string `json:"password"`
		RoleId   uint64 `json:"role_id"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	hashed, err := utils.HashPassword(input.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	user := entity.User{
		Name:      input.Name,
		Username:  input.Username,
		Password:  hashed,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.userRepo.Create(&user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Assign role
	if input.RoleId > 0 {
		_ = h.userRepo.AssignRole(user.ID, input.RoleId)
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *DashboardHandler) GetUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	user, err := h.userRepo.FindByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, user)
}

func (h *DashboardHandler) UpdateUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	existing, err := h.userRepo.FindByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	var input struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Password string `json:"password"`
		RoleId   uint64 `json:"role_id"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	existing.Name = input.Name
	existing.Username = input.Username
	if input.Password != "" {
		hashed, err := utils.HashPassword(input.Password)
		if err == nil {
			existing.Password = hashed
		}
	}
	existing.UpdatedAt = time.Now()

	if err := h.userRepo.Update(existing); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Re-assign role
	if input.RoleId > 0 {
		_ = h.userRepo.ClearRoles(existing.ID)
		_ = h.userRepo.AssignRole(existing.ID, input.RoleId)
	}

	return c.JSON(http.StatusOK, existing)
}

func (h *DashboardHandler) DeleteUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	_ = h.userRepo.ClearRoles(id)
	if err := h.userRepo.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

func (h *DashboardHandler) GetRoles(c echo.Context) error {
	roles, err := h.roleRepo.FindAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, roles)
}
