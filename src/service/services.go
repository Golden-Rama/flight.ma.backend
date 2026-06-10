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

	"gr-flight-ma-new/src/dto"
	"gr-flight-ma-new/src/entity"
	"gr-flight-ma-new/src/repository"
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

	for r := range resultChan {
		if r.err != nil {
			return nil, http.StatusInternalServerError, r.err
		}

		if r.statusCode == http.StatusUnauthorized {
			return nil, http.StatusUnauthorized, errors.New("unauthorized by flight provider")
		}

		if r.statusCode == http.StatusOK {
			var result dto.FlightSearchResponse
			if err := json.Unmarshal(r.res, &result); err != nil {
				continue
			}

			if len(result.Schedules) > 0 {
				gotDepartures = true
				for _, f := range result.Schedules[0].Flights {
					f.Provider = r.flightCode
					scheduleDepartures.Flights = append(scheduleDepartures.Flights, f)
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
						f.Provider = r.flightCode
						scheduleReturns.Flights = append(scheduleReturns.Flights, f)
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
	response.AirportDetails = []dto.AirportDetailResponse{}
	response.AirportV2Details = []dto.AirportV2DetailResponse{}
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
