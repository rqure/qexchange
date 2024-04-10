package main

import qmq "github.com/rqure/qmq/src"

type ProducerFactory struct {
	Config *Config
}

func (pf *ProducerFactory) Create(k string, p qmq.EngineComponentProvider) qmq.Producer {
	conn := p.WithConnectionProvider().Get("redis").(*qmq.RedisConnection)

	for _, producers := range pf.Config.ExchangeMap {
		for _, producer := range producers {
			if producer.Queue == k {
				return qmq.NewRedisProducer(k, conn, producer.Length, p.WithTransformerProvider().Get("producer:"+k))
			}
		}
	}

	return qmq.NewRedisProducer(k, conn, 10, p.WithTransformerProvider().Get("producer:"+k))
}
