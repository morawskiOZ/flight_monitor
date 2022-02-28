package signalObserver

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type Client struct {
	ExitChanel   chan int
	SignalChanel chan bool
}

func (c *Client) Observe() {
	osSignalChanel := make(chan os.Signal, 1)
	signal.Notify(osSignalChanel, os.Interrupt, syscall.SIGQUIT)

	var exitCode int
	s := <-osSignalChanel

syscallSwitch:
	switch s {
	case syscall.SIGQUIT:
		fmt.Println("Signal quit triggered.")
		break syscallSwitch
	default:
		fmt.Println("Unknown signal.")
		exitCode = 1
		break syscallSwitch
	}

	close(osSignalChanel)

	c.SignalChanel <- true
	c.ExitChanel <- exitCode
}

func NewOsSignalObserver() *Client {
	return &Client{
		ExitChanel:   make(chan int),
		SignalChanel: make(chan bool),
	}
}
