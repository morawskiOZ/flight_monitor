package main

import (
	"fmt"
	"os"
	"time"

	"github.com/morawskiOz/flight_monitor/configs"
	"github.com/morawskiOz/flight_monitor/pkg/mail"

	"github.com/morawskiOz/flight_monitor/internal/priceChecker"
	"github.com/morawskiOz/flight_monitor/internal/signal"
)

func main() {
	config, err := configs.LoadEnvConfig()
	if err != nil {
		fmt.Printf("Fatal error: %+v\n", err)
		os.Exit(1)
	}

	mc := mail.NewMailClient(
		mail.WithDialer(mail.MailAuthConfig{
			Port:         config.SmtpPort,
			Host:         config.SmtpHost,
			Password:     config.EmailPass,
			EmailAddress: config.EmailLogin,
		}),
	)

	so := signalObserver.NewOsSignalObserver()
	go so.Observe()

	pc := priceChecker.NewPriceChecker(priceChecker.WithMailClient(mc), priceChecker.WithSignalChannel(so.SignalChanel))
	t := []priceChecker.Task{
		{
			Destination: "SVQ",
			Origin:      "KRK",
			StartDate:   time.Date(2022, 02, 16, 0, 0, 0, 0, time.Now().Location()),
			EndDate:     time.Date(2022, 02, 23, 0, 0, 0, 0, time.Now().Location()),
			Treshold:    400,
			Recipient:   config.EmailRecipient,
			Subject:     "Sevilla route monitor",
		},
		{
			Destination: "MLA",
			Origin:      "WRO",
			StartDate:   time.Date(2022, 03, 5, 0, 0, 0, 0, time.Now().Location()),
			EndDate:     time.Date(2022, 03, 8, 0, 0, 0, 0, time.Now().Location()),
			Treshold:    200,
			Recipient:   config.EmailRecipient,
			Subject:     "Malta route monitor",
		},
	}

	pc.Run(t)

	exitCode := <-so.ExitChanel
	os.Exit(exitCode)
}
