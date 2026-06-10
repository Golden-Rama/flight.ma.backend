package dto

type ClassObjectResponse struct {
	Id             string `json:"Id"`
	FlightId       string `json:"FlightId"`
	Code           string `json:"Code"`
	Category       string `json:"Category"`
	Seat           any    `json:"Seat"`
	Fare           any    `json:"Fare"`
	Tax            any    `json:"Tax"`
	FareBasisCode  string `json:"FareBasisCode"`
	FareRuleKeys   string `json:"FareRuleKeys"`
	SegmentSellKey string `json:"SegmentSellKey"`
	ExtraData      string `json:"ExtraData"`
	Variant        any    `json:"Variant"`
	ClassGroupId   string `json:"ClassGroupId"`
	TransitTime    string `json:"TransitTime"`
	FareDesign     string `json:"FareDesign"`
}

type ConnectingFlightResponse FlightsResponse

type Facilities struct {
	Category    string `json:"Category"`
	Value       string `json:"Value"`
	Description string `json:"Description"`
	PaxType     string `json:"PaxType"`
}

type FareBreakdowns struct {
	PaxType string `json:"PaxType"`
	Charges []struct {
		Code            string `json:"Code"`
		Text            string `json:"Text"`
		Amount          int    `json:"Amount"`
		Currency        string `json:"Currency"`
		ForeignAmount   int    `json:"ForeignAmount"`
		ForeignCurrency string `json:"ForeignCurrency"`
	} `json:"Charges"`
}

type FlightsResponse struct {
	Id                       string                     `json:"Id"`
	GroupingId               string                     `json:"GroupingId"`
	Airline                  any                        `json:"Airline"`
	AirlineImageUrl          string                     `json:"AirlineImageUrl"`
	AirlineName              string                     `json:"AirlineName"`
	Number                   string                     `json:"Number"`
	OperatingNumber          string                     `json:"OperatingNumber"`
	OperatingAirlineImageUrl string                     `json:"OperatingAirlineImageUrl"`
	OperatingAirlineName     string                     `json:"OperatingAirlineName"`
	Origin                   string                     `json:"Origin"`
	OriginTerminal           string                     `json:"OriginTerminal"`
	Destination              string                     `json:"Destination"`
	DestinationTerminal      string                     `json:"DestinationTerminal"`
	Fare                     float32                    `json:"Fare"`
	FareType                 string                     `json:"FareType"`
	IsMultiClass             bool                       `json:"IsMultiClass"`
	IsConnecting             bool                       `json:"IsConnecting"`
	IsAvailable              bool                       `json:"IsAvailable"`
	FlightType               string                     `json:"FlightType"`
	DepartDate               string                     `json:"DepartDate"`
	DepartTime               string                     `json:"DepartTime"`
	ArriveDate               string                     `json:"ArriveDate"`
	ArriveTime               string                     `json:"ArriveTime"`
	Duration                 string                     `json:"Duration"`
	TotalTransit             int                        `json:"TotalTransit"`
	ClassObjects             []ClassObjectResponse      `json:"ClassObjects"`
	ConnectingFlights        []ConnectingFlightResponse `json:"ConnectingFlights"`
	FareBreakdowns           []FareBreakdowns           `json:"FareBreakdowns"`
	Facilities               []Facilities               `json:"Facilities"`
	PriceFare                string                     `json:"PriceFare"`
	IsOpenJawTransit         bool                       `json:"IsOpenJawTransit"`
	Note                     string                     `json:"Note"`
	Aircraft                 struct {
		IataCode  string `json:"IataCode"`
		ShortName string `json:"ShortName"`
		LongName  string `json:"LongName"`
	} `json:"Aircraft"`
	ExtraData              string `json:"ExtraData"`
	OriginAirportName      string `json:"OriginAirportName"`
	OriginCityName         string `json:"OriginCityName"`
	DepartDateTime         string `json:"DepartDateTime"`
	DestinationAirportName string `json:"DestinationAirportName"`
	DestinationCityName    string `json:"DestinationCityName"`
	ArriveDateTime         string `json:"ArriveDateTime"`
	TotalTransitTime       string `json:"TotalTransitTime"`
	TotalDateTime          string `json:"TotalDateTime"`
	Provider               string `json:"Provider,omitempty"`
}

type FlightSchedulesResponse struct {
	Origin                 string            `json:"Origin" example:"CGK"`
	Destination            string            `json:"Destination" example:"SIN"`
	IsInternational        bool              `json:"IsInternational" example:"true"`
	Flights                []FlightsResponse `json:"Flights"`
	OriginAirportName      string            `json:"OriginAirportName" example:"Soekarno-Hatta International Airport"`
	OriginCityName         string            `json:"OriginCityName" example:"Jakarta"`
	DestinationAirportName string            `json:"DestinationAirportName" example:"Changi Airport"`
	DestinationCityName    string            `json:"DestinationCityName" example:"Singapore"`
	Kind                   string            `json:"Kind" example:"CGK"`
}

type AirportDetailResponse struct {
	Code        string `json:"Code"`
	CountryCode string `json:"CountryCode"`
	CityName    string `json:"CityName"`
	AirportName string `json:"AirportName"`
	Locale      int    `json:"Locale"`
	Active      bool   `json:"Active"`
	CountryName string `json:"CountryName"`
	IsNonGds    bool   `json:"IsNonGds"`
	IsGds       bool   `json:"IsGds"`
	ForOverride string `json:"ForOverride"`
	LocalView   string `json:"LocalView"`
}

type AirportV2DetailResponse struct {
	Iata        string  `json:"Iata"`
	Name        string  `json:"Name"`
	City        string  `json:"City"`
	Country     string  `json:"Country"`
	CountryName string  `json:"CountryName"`
	IsGds       bool    `json:"IsGds"`
	IsNonGds    bool    `json:"IsNonGds"`
	Latitude    float32 `json:"Latitude"`
	Longitude   float32 `json:"Longitude"`
	TimeZoneZdb string  `json:"TimeZoneZdb"`
	MinOffset   string  `json:"MinOffset"`
	MaxOffset   string  `json:"MaxOffset"`
}

type FlightSearchResponse struct {
	Schedules        []FlightSchedulesResponse `json:"Schedules"`
	AirportDetails   []AirportDetailResponse   `json:"AirportDetails"`
	AirportV2Details []AirportV2DetailResponse `json:"AirportV2Details"`
}
