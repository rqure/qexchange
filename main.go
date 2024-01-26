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

type ProducerConfig struct {
	Queue string
	Length int64
}

func ParseExchangeMap() map[string][]ProducerConfig {
	var result map[string][][2]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		log.Fatalf("Error parsing JSON: %s", err)
	}
	
	// Convert to more structured data
	structuredResult := make(map[string][]ProducerConfig)
	for key, value := range result {
		for _, v := range value {
			queueName, ok := v[0].(string)
			if !ok {
				continue
			}
			
			number, ok := v[1].(float64) // JSON numbers are floats
			if !ok {
				continue
			}
			
			structuredResult[key] = append(structuredResult[key], ProducerConfig{Queue: queueName, Length: int64(number)})
		}
	}

	return structuredResult
}

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
