package main

import qmq "github.com/rqure/qmq/src"

type NameProvider struct{}

func (np *NameProvider) Get() string {
	return "exchange"
}

func main() {
	config := NewConfigParser()

	engine := qmq.NewDefaultEngine(qmq.DefaultEngineConfig{
		NameProvider:    &NameProvider{},
		EngineProcessor: &ExchangeEngineProcessor{Config: config},
		ProducerFactory: &ProducerFactory{Config: config},
	})
	engine.Run()
}
