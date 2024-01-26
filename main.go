package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	qmq "github.com/rqure/qmq/src"
)

type ProducerConfig struct {
	Queue  string
	Length int64
}

func ParseExchangeMap(configFile string, logger *qmq.QMQLogger) map[string][]ProducerConfig {
	var result map[string][][2]interface{}
	structuredResult := make(map[string][]ProducerConfig)

	content, err := os.ReadFile(configFile)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading file: %s", err))
		return structuredResult
	}

	err = json.Unmarshal(content, &result)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing JSON: %s", err))
		return structuredResult
	}

	for key, value := range result {
		for _, v := range value {
			queueName, ok := v[0].(string)
			if !ok {
				logger.Error(fmt.Sprintf("Encountered queue name that is not a string"))
				continue
			}

			length, ok := v[1].(float64) // JSON numbers are floats
			if !ok {
				logger.Error(fmt.Sprintf("Encountered length that is not a number"))
				continue
			}

			structuredResult[key] = append(structuredResult[key], ProducerConfig{Queue: queueName, Length: int64(length)})
			logger.Advise(fmt.Sprintf("Added binding between exchange '%s' and queue '%s' with up to %d entries", key, queueName, length))
		}
	}

	return structuredResult
}

func MainLoop(ctx context.Context, consumer string, producers []ProducerConfig, app *qmq.QMQApplication) {
	tickRateMs, err := strconv.Atoi(os.Getenv("TICK_RATE_MS"))
	if err != nil {
		tickRateMs = 100
	}

	ticker := time.NewTicker(time.Duration(tickRateMs) * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for {
				data, popped := app.Consumer(consumer).PopRaw()
				if popped == nil {
					break
				}

				app.Logger().Debug(fmt.Sprintf("Pushing %d bytes on exchange %s to %d queues", len(data), consumer, len(producers)))

				for _, producer := range producers {
					app.Producer(producer.Queue).PushRaw(data)
				}

				popped.Ack()
			}
		}
	}
}

func main() {
	app := qmq.NewQMQApplication("exchange")
	app.Initialize()
	defer app.Deinitialize()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	exchangeMap := ParseExchangeMap("exchanges.json", app.Logger())
	for consumer, producers := range exchangeMap {
		app.AddConsumer(consumer).Initialize()

		for _, producerConfig := range producers {
			app.AddProducer(producerConfig.Queue).Initialize(producerConfig.Length)
		}

		wg.Add(1)
		go func(ctx context.Context, consumer string, producers []ProducerConfig, app *qmq.QMQApplication) {
			defer wg.Done()
			MainLoop(ctx, consumer, producers, app)
		}(ctx, consumer, producers, app)
	}

	<-sigint
	app.Logger().Advise("SIGINT received")

	cancel()
	wg.Wait()
}
