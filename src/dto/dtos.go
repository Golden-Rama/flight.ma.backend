package dto

type FlightSearchInput struct {
	DepartureDate     string   `json:"departureDate" validate:"required"`
	ReturnDate        string   `json:"returnDate,omitempty"`
	OriginCode        string   `json:"originCode" validate:"required"`
	DestinationCode   string   `json:"destinationCode" validate:"required"`
	IsRoundTrip       string   `json:"isRoundTrip" validate:"required"`
	Adult             any      `json:"adult" validate:"required"`
	Child             any      `json:"child" validate:"required"`
	Infant            any      `json:"infant" validate:"required"`
	PreferredCarriers []string `json:"preferredCarriers,omitempty"`
	CabinClass        string   `json:"cabinClass,omitempty"`
	CabinClasses      []string `json:"cabinClasses,omitempty"`
	FlightType        string   `json:"flightType" validate:"required"`
}

type PassengerInput struct {
	Index             any           `json:"index"`
	Type              any           `json:"type"`
	Title             string        `json:"title"`
	FirstName         string        `json:"firstName"`
	LastName          string        `json:"lastName"`
	IsSeniorCitizen   string        `json:"isSeniorCitizen"`
	BirthDate         string        `json:"birthDate"`
	Email             string        `json:"email"`
	HomePhone         string        `json:"homePhone"`
	MobilePhone       string        `json:"mobilePhone"`
	OtherPhone        string        `json:"otherPhone"`
	IdNumber          string        `json:"idNumber"`
	Nationality       string        `json:"nationality"`
	AdultAssoc        string        `json:"adultAssoc"`
	PassportNumber    string        `json:"passportNumber"`
	PassportExpire    string        `json:"passportExpire"`
	PassportOrigin    string        `json:"passportOrigin"`
	EmergencyFullName string        `json:"emergencyFullName"`
	EmergencyPhone    string        `json:"emergencyPhone"`
	EmergencyEmail    string        `json:"emergencyEmail"`
	Seats             []interface{} `json:"seats"`
	Ssrs              []interface{} `json:"ssrs"`
}

type SegmentInput struct {
	ClassId      string `json:"classId"`
	Airline      any    `json:"Airline"`
	FlightNumber string `json:"flightNumber"`
	Origin       string `json:"origin"`
	DepartDate   string `json:"departDate"`
	DepartTime   string `json:"departTime"`
	Destination  string `json:"destination"`
	ArriveDate   string `json:"arriveDate"`
	ArriveTime   string `json:"arriveTime"`
	ClassCode    string `json:"classCode"`
	FlightId     string `json:"flightId"`
	Num          any    `json:"num"`
	Seq          any    `json:"seq"`
}

type SelectedOfferInput struct {
	OfferItemId    string `json:"offerItemId"`
	IndexUser      []int  `json:"indexUser"`
	PaxSegmentType string `json:"paxSegmentType"`
	PaxSegmentRef  string `json:"paxSegmentRef"`
}

type SelectedSeatInput struct {
	OfferItemId   string `json:"offerItemId"`
	IndexUser     int    `json:"index"`
	PaxSegmentRef string `json:"paxSegmentRef"`
	ColumnId      string `json:"columnId"`
	SeatRowNumber int    `json:"seatRowNumber"`
}

type OfferInput struct {
	OfferId   string               `json:"offerId"`
	OwnerCode string               `json:"ownerCode"`
	Selected  []SelectedOfferInput `json:"selectedOffer"`
}

type SeatInput struct {
	OfferId   string              `json:"offerId"`
	OwnerCode string              `json:"ownerCode"`
	Selected  []SelectedSeatInput `json:"selectedSeat"`
}

type FlightBookingInput struct {
	AvailableDomestic string           `json:"availableDomestic"`
	Contact           ContactInput     `json:"contact"`
	Passengers        []PassengerInput `json:"passengers"`
	Segments          []SegmentInput   `json:"segments"`
	CallbackUri       string           `json:"callbackUri"`
	FlightType        string           `json:"flightType"`
	Offers            *[]OfferInput    `json:"offers,omitempty"`
	Seats             *[]SeatInput     `json:"seats,omitempty"`
}

type ContactInput struct {
	Email       string `json:"email"`
	Title       string `json:"title"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	HomePhone   string `json:"homePhone"`
	MobilePhone string `json:"mobilePhone"`
}

type PnrInput struct {
	PnrID string `json:"pnrId" validate:"required"`
}

type FareDetailInput struct {
	AvailableDomestic string         `json:"availableDomestic" validate:"required"`
	Contact           ContactInput   `json:"contact" validate:"required"`
	Passengers        PassengerInput `json:"passengers" validate:"required"`
	Segments          []SegmentInput `json:"segments" validate:"required"`
	FlightType        string         `json:"flightType" validate:"required"`
}
