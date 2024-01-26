package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	qmq "github.com/rqure/qmq/src"
)

func main() {
	app := qmq.NewQMQApplication("exchange")
	app.Initialize()
	defer app.Deinitialize()

	key := os.Getenv("APP_NAME") + ":logs"
	app.AddConsumer(key).Initialize()
	app.Consumer(key).ResetLastId()

	tickRateMs, err := strconv.Atoi(os.Getenv("TICK_RATE_MS"))
	if err != nil {
		tickRateMs = 100
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	ticker := time.NewTicker(time.Duration(tickRateMs) * time.Millisecond)
	for {
		select {
		case <-sigint:
			return
		case <-ticker.C:

		}
	}
}
