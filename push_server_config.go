// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
)

// PushServerConfig represents Redis connection information
type PushServerConfig struct {
	Addresses     []string
	DefaultTopic  string
	Authorizer    Authorizer
	Authenticator Authenticator
}

// NewPushServerConfig returns a new RedisInfo
func NewPushServerConfig(addresses []string, defaultTopic string) *PushServerConfig {

	if len(addresses) < 1 {
		panic("at least one address should be provided to PushServerConfig")
	}

	if defaultTopic == "" {
		panic("a valid default topic should be provided to PushServerConfig")
	}

	return &PushServerConfig{
		Addresses:    addresses,
		DefaultTopic: defaultTopic,
	}
}

func (k *PushServerConfig) makeProducer() sarama.SyncProducer {

	producer, err := sarama.NewSyncProducer(k.Addresses, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"info":  k,
			"error": err,
		}).Error("unable to create kafka producer")

		return nil
	}

	return producer
}

func (k *PushServerConfig) makeConsumer() sarama.Consumer {

	consumer, err := sarama.NewConsumer(k.Addresses, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"info":  k,
			"error": err,
		}).Error("unable to create kafka consumer")

		return nil
	}

	return consumer
}

func (k *PushServerConfig) String() string {

	return fmt.Sprintf("<PushServerConfig Addresses: %v DefaultTopic: %s>", k.Addresses, k.DefaultTopic)
}
