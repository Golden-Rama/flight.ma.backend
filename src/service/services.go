package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"flight.ma.backend/src/dto"
	"flight.ma.backend/src/entity"
	"flight.ma.backend/src/repository"
)

type SearchService interface {
	Search(ctx context.Context, input dto.FlightSearchInput, authHeader string) (*dto.FlightSearchResponse, int, error)
}

type mysqlSearchService struct {
	providerRepo repository.ProviderRepository
	client       *http.Client
}

func NewSearchService(providerRepo repository.ProviderRepository) SearchService {
	return &mysqlSearchService{
		providerRepo: providerRepo,
		client: &http.Client{
			Timeout: 1 * time.Minute,
		},
	}
}

type searchResult struct {
	res        []byte
	statusCode int
	err        error
	flightCode string
}

func (s *mysqlSearchService) Search(ctx context.Context, input dto.FlightSearchInput, authHeader string) (*dto.FlightSearchResponse, int, error) {
	// Find all active providers
	providers, err := s.providerRepo.FindAll(true)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if len(providers) == 0 {
		return nil, http.StatusNotFound, errors.New("no flight providers available")
	}

	var wg sync.WaitGroup
	resultChan := make(chan searchResult, len(providers))

	for _, provider := range providers {
		if provider.FlightQuest == nil {
			continue
		}

		// Filter matching flight type
		if (input.FlightType == "NonGDS" && provider.AvailableDomestic) || (input.FlightType == "GDS") {
			wg.Add(1)

			go func(p entity.FlightProvider) {
				defer wg.Done()

				// Setup preferred carriers
				preferredCarriers := []string{}
				if p.FlightQuest.PreferredCarriersGds != "" {
					preferredCarriers = append(preferredCarriers, strings.Split(p.FlightQuest.PreferredCarriersGds, ",")...)
				}
				if p.FlightQuest.PreferredCarriersNonGds != "" {
					preferredCarriers = append(preferredCarriers, strings.Split(p.FlightQuest.PreferredCarriersNonGds, ",")...)
				}

				providerInput := input
				providerInput.PreferredCarriers = preferredCarriers

				payload, err := json.Marshal(providerInput)
				if err != nil {
					resultChan <- searchResult{err: err, flightCode: p.Code}
					return
				}

				targetUrl := strings.ReplaceAll(p.BaseUrl+p.FlightQuest.Endpoint, " ", "")
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetUrl, bytes.NewBuffer(payload))
				if err != nil {
					resultChan <- searchResult{err: err, flightCode: p.Code}
					return
				}

				req.Header.Set("Content-Type", "application/json")
				if authHeader != "" {
					req.Header.Set("Authorization", authHeader)
				}

				resp, err := s.client.Do(req)
				if err != nil {
					resultChan <- searchResult{err: err, flightCode: p.Code}
					return
				}
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					resultChan <- searchResult{err: err, flightCode: p.Code}
					return
				}

				resultChan <- searchResult{
					res:        body,
					statusCode: resp.StatusCode,
					flightCode: p.Code,
				}
			}(provider)
		}
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var scheduleDepartures dto.FlightSchedulesResponse
	var scheduleReturns dto.FlightSchedulesResponse
	var gotDepartures, gotReturns bool

	airportMap := make(map[string]dto.AirportDetailResponse)
	airportV2Map := make(map[string]dto.AirportV2DetailResponse)

	for r := range resultChan {
		if r.err != nil {
			return nil, http.StatusInternalServerError, r.err
		}

		if r.statusCode == http.StatusUnauthorized {
			return nil, http.StatusUnauthorized, errors.New("unauthorized by flight provider")
		}

		if r.statusCode == http.StatusOK {
			var result dto.FlightSearchResponse
			var parsed bool

			// Parse response, handling various wrapping formats (double-wrapped, single-wrapped, unwrapped)
			type genericMap map[string]json.RawMessage
			var top genericMap
			if err := json.Unmarshal(r.res, &top); err == nil {
				// Case 1: Double-wrapped or Single-wrapped in "data"
				if dataRaw, ok := top["data"]; ok {
					var inner genericMap
					if err := json.Unmarshal(dataRaw, &inner); err == nil {
						// Double-wrapped: {"success": true, "data": {"status": true, "data": {...}}}
						if innerDataRaw, ok := inner["data"]; ok {
							if err := json.Unmarshal(innerDataRaw, &result); err == nil && (len(result.Schedules) > 0 || len(result.AirportDetails) > 0) {
								parsed = true
							}
						}
					}
					// Single-wrapped in "data": {"success": true, "data": {...}}
					if !parsed {
						if err := json.Unmarshal(dataRaw, &result); err == nil && (len(result.Schedules) > 0 || len(result.AirportDetails) > 0) {
							parsed = true
						}
					}
				}
				// Case 2: Single-wrapped with status/data: {"status": true, "data": {...}}
				if !parsed {
					if dataRaw, ok := top["data"]; ok {
						if err := json.Unmarshal(dataRaw, &result); err == nil && (len(result.Schedules) > 0 || len(result.AirportDetails) > 0) {
							parsed = true
						}
					}
				}
			}

			// Case 3: Unwrapped response: {...}
			if !parsed {
				if err := json.Unmarshal(r.res, &result); err == nil {
					parsed = true
				}
			}

			if !parsed {
				continue
			}

			// Find provider configuration
			var currentProvider entity.FlightProvider
			for _, p := range providers {
				if p.Code == r.flightCode {
					currentProvider = p
					break
				}
			}

			preferredGds := make(map[string]bool)
			preferredNonGds := make(map[string]bool)
			if currentProvider.FlightQuest != nil {
				if currentProvider.FlightQuest.PreferredCarriersGds != "" {
					for _, code := range strings.Split(currentProvider.FlightQuest.PreferredCarriersGds, ",") {
						cCode := strings.TrimSpace(strings.ToUpper(code))
						if cCode != "" {
							preferredGds[cCode] = true
						}
					}
				}
				if currentProvider.FlightQuest.PreferredCarriersNonGds != "" {
					for _, code := range strings.Split(currentProvider.FlightQuest.PreferredCarriersNonGds, ",") {
						cCode := strings.TrimSpace(strings.ToUpper(code))
						if cCode != "" {
							preferredNonGds[cCode] = true
						}
					}
				}
			}

			isPreferredFlight := func(f dto.FlightsResponse) bool {
				carrierCode := ""
				if len(f.Number) >= 2 {
					carrierCode = strings.ToUpper(f.Number[:2])
				}
				if carrierCode == "" {
					return true
				}

				isGds := strings.EqualFold(f.FlightType, "Gds")

				if isGds {
					if len(preferredGds) > 0 {
						return preferredGds[carrierCode]
					}
				} else {
					if len(preferredNonGds) > 0 {
						return preferredNonGds[carrierCode]
					}
				}
				return true
			}

			allowedAirports := make(map[string]bool)

			if len(result.Schedules) > 0 {
				gotDepartures = true
				for _, f := range result.Schedules[0].Flights {
					if !isPreferredFlight(f) {
						continue
					}
					f.Provider = r.flightCode
					if f.AirlineName == "" && len(f.ConnectingFlights) > 0 {
						f.AirlineName = f.ConnectingFlights[0].AirlineName
					}
					if f.AirlineImageUrl == "" && len(f.ConnectingFlights) > 0 {
						f.AirlineImageUrl = f.ConnectingFlights[0].AirlineImageUrl
					}
					scheduleDepartures.Flights = append(scheduleDepartures.Flights, f)

					// Mark airports used in this allowed flight
					allowedAirports[strings.ToUpper(f.Origin)] = true
					allowedAirports[strings.ToUpper(f.Destination)] = true
					for _, cf := range f.ConnectingFlights {
						allowedAirports[strings.ToUpper(cf.Origin)] = true
						allowedAirports[strings.ToUpper(cf.Destination)] = true
					}
				}
				scheduleDepartures.Origin = result.Schedules[0].Origin
				scheduleDepartures.Destination = result.Schedules[0].Destination
				scheduleDepartures.IsInternational = result.Schedules[0].IsInternational
				scheduleDepartures.OriginAirportName = result.Schedules[0].OriginAirportName
				scheduleDepartures.OriginCityName = result.Schedules[0].OriginCityName
				scheduleDepartures.DestinationAirportName = result.Schedules[0].DestinationAirportName
				scheduleDepartures.DestinationCityName = result.Schedules[0].DestinationCityName
				scheduleDepartures.Kind = "Departure"

				if input.IsRoundTrip == "true" && len(result.Schedules) > 1 {
					gotReturns = true
					for _, f := range result.Schedules[1].Flights {
						if !isPreferredFlight(f) {
							continue
						}
						f.Provider = r.flightCode
						if f.AirlineName == "" && len(f.ConnectingFlights) > 0 {
							f.AirlineName = f.ConnectingFlights[0].AirlineName
						}
						if f.AirlineImageUrl == "" && len(f.ConnectingFlights) > 0 {
							f.AirlineImageUrl = f.ConnectingFlights[0].AirlineImageUrl
						}
						scheduleReturns.Flights = append(scheduleReturns.Flights, f)

						// Mark airports used in this allowed flight
						allowedAirports[strings.ToUpper(f.Origin)] = true
						allowedAirports[strings.ToUpper(f.Destination)] = true
						for _, cf := range f.ConnectingFlights {
							allowedAirports[strings.ToUpper(cf.Origin)] = true
							allowedAirports[strings.ToUpper(cf.Destination)] = true
						}
					}
					scheduleReturns.Origin = result.Schedules[1].Origin
					scheduleReturns.Destination = result.Schedules[1].Destination
					scheduleReturns.IsInternational = result.Schedules[1].IsInternational
					scheduleReturns.OriginAirportName = result.Schedules[1].OriginAirportName
					scheduleReturns.OriginCityName = result.Schedules[1].OriginCityName
					scheduleReturns.DestinationAirportName = result.Schedules[1].DestinationAirportName
					scheduleReturns.DestinationCityName = result.Schedules[1].DestinationCityName
					scheduleReturns.Kind = "Return"
				}
			}

			// Aggregate airport details only if they are used by allowed flights
			for _, port := range result.AirportDetails {
				if allowedAirports[strings.ToUpper(port.Code)] {
					airportMap[port.Code] = port
				}
			}
			for _, port := range result.AirportV2Details {
				if allowedAirports[strings.ToUpper(port.Iata)] {
					airportV2Map[port.Iata] = port
				}
			}
		}
	}

	// Sort and filter duplicate flight schedules
	sort.Slice(scheduleDepartures.Flights, func(i, j int) bool {
		return scheduleDepartures.Flights[i].Fare > scheduleDepartures.Flights[j].Fare
	})
	scheduleDepartures.Flights = uniqueFlights(scheduleDepartures.Flights)

	sort.Slice(scheduleReturns.Flights, func(i, j int) bool {
		return scheduleReturns.Flights[i].Fare > scheduleReturns.Flights[j].Fare
	})
	scheduleReturns.Flights = uniqueFlights(scheduleReturns.Flights)

	var response dto.FlightSearchResponse

	// Collect unique AirportDetails and sort them alphabetically
	response.AirportDetails = []dto.AirportDetailResponse{}
	for _, port := range airportMap {
		response.AirportDetails = append(response.AirportDetails, port)
	}
	sort.Slice(response.AirportDetails, func(i, j int) bool {
		return response.AirportDetails[i].Code < response.AirportDetails[j].Code
	})

	// Collect unique AirportV2Details and sort them alphabetically
	response.AirportV2Details = []dto.AirportV2DetailResponse{}
	for _, port := range airportV2Map {
		response.AirportV2Details = append(response.AirportV2Details, port)
	}
	sort.Slice(response.AirportV2Details, func(i, j int) bool {
		return response.AirportV2Details[i].Iata < response.AirportV2Details[j].Iata
	})

	response.Schedules = []dto.FlightSchedulesResponse{}

	if gotDepartures && len(scheduleDepartures.Flights) > 0 {
		response.Schedules = append(response.Schedules, scheduleDepartures)
	}
	if gotReturns && len(scheduleReturns.Flights) > 0 {
		response.Schedules = append(response.Schedules, scheduleReturns)
	}

	return &response, http.StatusOK, nil
}

func uniqueFlights(data []dto.FlightsResponse) []dto.FlightsResponse {
	var unique []dto.FlightsResponse
	type key struct{ val string }
	m := make(map[key]int)
	for _, v := range data {
		if v.AirlineName == "Singapore Airlines" {
			unique = append(unique, v)
		} else {
			k := key{val: v.DepartTime}
			if i, ok := m[k]; ok {
				unique[i] = v
			} else {
				m[k] = len(unique)
				unique = append(unique, v)
			}
		}
	}
	return unique
}

// BookingService interface
type BookingService interface {
	ProxyRequest(ctx context.Context, providerCode string, serviceType string, body []byte, authHeader string) ([]byte, int, error)
}

type mysqlBookingService struct {
	providerRepo repository.ProviderRepository
	client       *http.Client
}

func NewBookingService(providerRepo repository.ProviderRepository) BookingService {
	return &mysqlBookingService{
		providerRepo: providerRepo,
		client: &http.Client{
			Timeout: 1 * time.Minute,
		},
	}
}

func (s *mysqlBookingService) ProxyRequest(ctx context.Context, providerCode string, serviceType string, body []byte, authHeader string) ([]byte, int, error) {
	provider, err := s.providerRepo.FindByCode(providerCode)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("provider %s is not registered", providerCode)
	}

	if provider.FlightBooking == nil {
		return nil, http.StatusBadRequest, fmt.Errorf("booking endpoints not configured for provider %s", providerCode)
	}

	endpoint := ""
	switch serviceType {
	case "fare-detail":
		endpoint = provider.FlightBooking.EndpointFareDetail
	case "reservation":
		endpoint = provider.FlightBooking.EndpointBooking
	case "check-reservation":
		endpoint = provider.FlightBooking.EndpointCheckBooking
	case "issue-ticket":
		endpoint = provider.FlightBooking.EndpointIssueTicket
	case "cancel-reservation":
		endpoint = provider.FlightBooking.EndpointCancelBooking
	}

	if endpoint == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("endpoint for %s is not defined", serviceType)
	}

	targetUrl := strings.ReplaceAll(provider.BaseUrl+endpoint, " ", "")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return resBody, resp.StatusCode, nil
}
