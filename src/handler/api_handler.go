package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"flight.ma.backend/src/dto"
	"flight.ma.backend/src/service"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type ApiHandler struct {
	searchService  service.SearchService
	bookingService service.BookingService
	validate       *validator.Validate
}

func NewApiHandler(searchService service.SearchService, bookingService service.BookingService) *ApiHandler {
	return &ApiHandler{
		searchService:  searchService,
		bookingService: bookingService,
		validate:       validator.New(),
	}
}

func (h *ApiHandler) RegisterRoutes(e *echo.Echo, authMiddleware echo.MiddlewareFunc) {
	// Public oauth2 endpoint
	e.POST("/oauth2/token", h.OAuth2Token)

	// API Group requiring Bearer JWT Token
	api := e.Group("/api/v1", authMiddleware)
	api.POST("/search", h.Search)
	api.POST("/fare-detail", h.FareDetail)
	api.POST("/reservation", h.Reservation)
	api.POST("/check-reservation", h.CheckReservation)
	api.POST("/issue-ticket", h.IssueTicket)
	api.POST("/cancel-reservation", h.CancelReservation)
}

func (h *ApiHandler) OAuth2Token(c echo.Context) error {
	// Simple client credentials flow implementation matching flight.service standard
	var req struct {
		ClientID     string `json:"client_id" form:"client_id"`
		ClientSecret string `json:"client_secret" form:"client_secret"`
		GrantType    string `json:"grant_type" form:"grant_type"`
	}

	_ = c.Bind(&req)

	if req.GrantType != "client_credentials" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "unsupported_grant_type",
		})
	}

	if req.ClientID == "" || req.ClientSecret == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "invalid_client",
		})
	}

	// Sign a dummy JWT token for demo/standard compatibility
	// In production, we'd validate client ID and secret against DB
	token := "dummy_jwt_token_for_" + req.ClientID

	return c.JSON(http.StatusOK, map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}

func (h *ApiHandler) Search(c echo.Context) error {
	var input dto.FlightSearchInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	if err := h.validate.Struct(input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	data, code, err := h.searchService.Search(c.Request().Context(), input, authHeader)
	if err != nil {
		return c.JSON(code, map[string]any{"error": err.Error()})
	}

	return c.JSON(code, data)
}

func (h *ApiHandler) FareDetail(c echo.Context) error {
	return h.proxyBooking(c, "fare-detail")
}

func (h *ApiHandler) Reservation(c echo.Context) error {
	return h.proxyBooking(c, "reservation")
}

func (h *ApiHandler) CheckReservation(c echo.Context) error {
	return h.proxyBooking(c, "check-reservation")
}

func (h *ApiHandler) IssueTicket(c echo.Context) error {
	return h.proxyBooking(c, "issue-ticket")
}

func (h *ApiHandler) CancelReservation(c echo.Context) error {
	return h.proxyBooking(c, "cancel-reservation")
}

func (h *ApiHandler) proxyBooking(c echo.Context, serviceType string) error {
	providerCode := c.Request().Header.Get("X-Provider")
	if providerCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": "X-Provider header is required"})
	}

	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}

	authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
	res, code, err := h.bookingService.ProxyRequest(c.Request().Context(), providerCode, serviceType, bodyBytes, authHeader)
	if err != nil {
		return c.JSON(code, map[string]any{"error": err.Error()})
	}

	var jsonRes any
	if err := json.Unmarshal(res, &jsonRes); err == nil {
		return c.JSON(code, jsonRes)
	}

	return c.Blob(code, "application/json", res)
}
