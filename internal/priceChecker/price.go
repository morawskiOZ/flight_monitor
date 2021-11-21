package priceChecker

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/morawskiOz/flight_monitor/internal/ryanair"
	"github.com/morawskiOz/flight_monitor/pkg/mail"
)

type client struct {
	mailClient      *mail.Client
	wg              *sync.WaitGroup
	close           chan bool
	signalChannel   chan bool
	observableTasks []*observableTask
}

type Task struct {
	Destination string
	Origin      string
	StartDate   time.Time
	EndDate     time.Time
	Treshold    float64
	Recipient   string
	Subject     string
}

type PriceClient interface {
	GetPrice() (float64, error)
}

type observableTask struct {
	Task
	PriceClient
	checksCount int
	low         float64
	ticker      *time.Ticker
}

type option func(*client)

func WithMailClient(mc *mail.Client) option {
	return func(pc *client) {
		pc.mailClient = mc
	}
}

func WithSignalChannel(sch chan bool) option {
	return func(pc *client) {
		pc.signalChannel = sch
	}
}

func NewPriceChecker(options ...option) *client {
	c := &client{
		wg: &sync.WaitGroup{},
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

func (pc *client) scheduleTask(ot *observableTask) {
	for {
		select {
		case <-ot.ticker.C:
			pc.wg.Add(1)
			if err := pc.runTask(ot); err != nil {
				fmt.Printf("Error: %+v\n", err)
			}
		case <-pc.close:
			fmt.Println("Error: close signal, all tasks cancelled")
			return
		}
	}

}

func (pc *client) newObservableTask(t Task) *observableTask {
	priceClient := ryanair.NewRyanairClient(ryanair.WithPriceEndpoint(t.StartDate, t.EndDate, t.Destination, t.Origin))
	defaultLow := 9999.99
	ot := observableTask{
		ticker:      time.NewTicker(60 * time.Minute),
		Task:        t,
		checksCount: 0,
		low:         defaultLow,
		PriceClient: priceClient,
	}

	return &ot
}

func (pc *client) runTask(ot *observableTask) error {
	defer pc.wg.Done()

	price, err := ot.GetPrice()
	fmt.Println(price)
	if err != nil {
		return err
	}

	ot.checksCount = ot.checksCount + 1
	ot.low = math.Min(ot.low, price)

	if err := pc.sendNotification(ot, price); err != nil {
		return err
	}

	return nil
}

func (pc *client) sendNotification(ot *observableTask, price float64) error {
	if price < ot.Treshold {
		msg := fmt.Sprintf("Time to buy ticket on route: %v - %v at price: %v", ot.Origin, ot.Destination, price)
		if err := pc.mailClient.Send(ot.Recipient, ot.Subject, msg); err != nil {
			return err
		}

		return nil

	}
	if ot.checksCount%24 == 0 {
		msg := fmt.Sprintf("Route %v - %v is monitored, lowest price observed: %v", ot.Origin, ot.Destination, ot.low)
		if err := pc.mailClient.Send(ot.Recipient, ot.Subject, msg); err != nil {
			return err
		}
	}

	return nil
}

func (pc *client) Run(tasks []Task) {
	close := make(chan bool, len(tasks))
	pc.close = close

	for _, t := range tasks {
		ot := pc.newObservableTask(t)
		go pc.scheduleTask(ot)
		pc.observableTasks = append(pc.observableTasks, ot)
	}

	exit := <-pc.signalChannel
	if exit {
		for _, ot := range pc.observableTasks {
			ot.ticker.Stop()
			pc.close <- true

			if appEnv := os.Getenv("APP_ENV"); appEnv == "production" {
				pc.mailClient.Send(ot.Recipient, ot.Subject, "Price no longer observed, server is shouting down")
			}
		}
	}
	pc.wg.Wait()
}
