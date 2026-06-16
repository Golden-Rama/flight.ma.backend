package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"flight.ma.backend/src/dto"
	"flight.ma.backend/src/service"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type ApiHandler struct {
	searchService  service.SearchService
	bookingService service.BookingService
	jwtSecret      string
	validate       *validator.Validate
}

func NewApiHandler(searchService service.SearchService, bookingService service.BookingService, jwtSecret string) *ApiHandler {
	return &ApiHandler{
		searchService:  searchService,
		bookingService: bookingService,
		jwtSecret:      jwtSecret,
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

	// Sign a real JWT token using HS256 and the configured JwtSecret
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "failed_to_generate_token_header",
		})
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)

	expiresIn := int64(3600)
	payload := map[string]any{
		"sub": req.ClientID,
		"exp": time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
		"iat": time.Now().Unix(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "failed_to_generate_token_payload",
		})
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	signingInput := headerB64 + "." + payloadB64

	mac := hmac.New(sha256.New, []byte(h.jwtSecret))
	mac.Write([]byte(signingInput))
	signatureB64 := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	token := signingInput + "." + signatureB64

	return c.JSON(http.StatusOK, map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   expiresIn,
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
		if serviceType == "fare-detail" {
			if targetMap, found := findClassIdMap(jsonRes); found {
				return c.JSON(code, targetMap)
			}
		}
		return c.JSON(code, jsonRes)
	}

	return c.Blob(code, "application/json", res)
}

func findClassIdMap(val any) (map[string]any, bool) {
	m, ok := val.(map[string]any)
	if !ok {
		return nil, false
	}
	if _, exists := m["ClassId"]; exists {
		return m, true
	}
	if _, exists := m["classId"]; exists {
		return m, true
	}
	if dataVal, exists := m["data"]; exists {
		if res, ok := findClassIdMap(dataVal); ok {
			return res, true
		}
	}
	return nil, false
}
