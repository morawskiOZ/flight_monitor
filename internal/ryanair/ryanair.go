package ryanair

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	layoutISO = "2006-01-02"
)

var PriceNotAvailableErr = errors.New("Price not available for that route, check parameters")

type priceResponse struct {
	Fares []fares `json:"fares"`
}

type fares struct {
	Summary summary `json:"summary"`
}

type summaryPrice struct {
	Value          float64 `json:"value"`
	CurrencyCode   string  `json:"currencyCode"`
	CurrencySymbol string  `json:"currencySymbol"`
}

type summary struct {
	Price         summaryPrice `json:"price"`
	PreviousPrice summaryPrice `json:"previousPrice"`
}

type client struct {
	httpClient *http.Client
	url        string
}

type option func(*client)

func WithPriceEndpoint(startDate time.Time, endDate time.Time, destination string, origin string) option {
	sd := startDate.Format(layoutISO)
	ed := endDate.Format(layoutISO)
	url := fmt.Sprintf("https://www.ryanair.com/api/farfnd/3/roundTripFares?&arrivalAirportIataCode=%v&departureAirportIataCode=%v&inboundDepartureDateFrom=%v&inboundDepartureDateTo=%v&language=pl&limit=1&market=pl-pl&offset=0&outboundDepartureDateFrom=%v&outboundDepartureDateTo=%v", destination, origin, ed, ed, sd, sd)
	return func(c *client) {
		c.url = url
	}

}

func NewRyanairClient(options ...option) *client {
	client := &client{}
	hc := &http.Client{
		Timeout: time.Second * 10,
	}
	client.httpClient = hc

	for _, opt := range options {
		opt(client)
	}

	return client
}

func (c *client) GetPrice() (float64, error) {
	req, err := http.NewRequest("GET", c.url, nil)
	if err != nil {
		return 0, errors.Wrap(err, "Error creating request")
	}

	req.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "API request error")
	}

	defer response.Body.Close()

	res := priceResponse{}

	if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
		return 0, errors.Wrap(err, "Decoding API response error")
	}

	if len(res.Fares) == 0 {
		return 0, PriceNotAvailableErr
	}

	return res.Fares[0].Summary.Price.Value, nil
}
