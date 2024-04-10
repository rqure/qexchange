package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	qmq "github.com/rqure/qmq/src"
)

type ExchangeEngineProcessor struct {
	Config *Config
}

func (e *ExchangeEngineProcessor) Process(p qmq.EngineComponentProvider) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	e.Config.Parse("exchanges.json", p.WithLogger())

	for consumer := range e.Config.ExchangeMap {
		go e.Subprocess(consumer, p)
	}

	<-quit
}

func (e *ExchangeEngineProcessor) Subprocess(k string, p qmq.EngineComponentProvider) {
	producers := e.Config.ExchangeMap[k]

	for consumable := range p.WithConsumer(k).Pop() {
		consumable.Ack()
		message := consumable.Data()
		for _, producer := range producers {
			p.WithLogger().Debug(fmt.Sprintf("Pushing message on exchange %s to queue %s: %v", k, producer.Queue, message))
			p.WithProducer(producer.Queue).Push(message)
		}
	}
}
