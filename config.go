package main

import (
	"encoding/json"
	"fmt"
	"os"

	qmq "github.com/rqure/qmq/src"
)

type ProducerConfig struct {
	Queue  string
	Length int64
}

type Config struct {
	ExchangeMap map[string][]ProducerConfig
}

func NewConfigParser() *Config {
	return &Config{
		ExchangeMap: make(map[string][]ProducerConfig),
	}
}

func (c *Config) Parse(configFile string, logger qmq.Logger) {
	var result map[string][][2]interface{}
	structuredResult := make(map[string][]ProducerConfig)

	content, err := os.ReadFile(configFile)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading file: %s", err))
		c.ExchangeMap = structuredResult
		return
	}

	err = json.Unmarshal(content, &result)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing JSON: %s", err))
		c.ExchangeMap = structuredResult
		return
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
			logger.Advise(fmt.Sprintf("Added binding between exchange '%s' and queue '%s' with up to %d entries", key, queueName, int(length)))
		}
	}

	c.ExchangeMap = structuredResult
}
