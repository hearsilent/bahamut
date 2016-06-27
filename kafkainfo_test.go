// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"testing"

	"github.com/Shopify/sarama"
	. "github.com/smartystreets/goconvey/convey"
)

func TestKakfaInfo_NewKafkaInfo(t *testing.T) {

	Convey("Given I create have a new kafka info", t, func() {

		kafkaInfo := NewKafkaInfo([]string{":1234"}, "topic")

		Convey("Then the kafka info should have the address set", func() {
			So(len(kafkaInfo.Addresses), ShouldEqual, 1)
			So(kafkaInfo.Addresses[0], ShouldEqual, ":1234")
		})

		Convey("Then the kafka info should have the topic set", func() {
			So(kafkaInfo.Topic, ShouldEqual, "topic")
		})
	})

	Convey("Given I create have a new kafka info with an empty address array", t, func() {

		Convey("Then it should panic ", func() {
			So(func() { NewKafkaInfo([]string{}, "topic") }, ShouldPanic)
		})
	})

	Convey("Given I create have a new kafka info with an empty topic", t, func() {

		Convey("Then it should panic ", func() {
			So(func() { NewKafkaInfo([]string{":1234"}, "") }, ShouldPanic)
		})
	})
}

func TestKakfaInfo_String(t *testing.T) {

	Convey("Given I create have a new kafka info", t, func() {

		KafkaInfo := NewKafkaInfo([]string{"127.0.0.1:1234", "127.0.0.1:1235"}, "topic")

		Convey("Then the string representation should be correct", func() {
			So(KafkaInfo.String(), ShouldEqual, "<kafkaInfo addresses: [127.0.0.1:1234 127.0.0.1:1235] topic: topic>")
		})
	})
}

func TestKakfaInfo_makeProducer(t *testing.T) {

	Convey("Given I create have a new kafka info with a kafka server listen", t, func() {

		broker := sarama.NewMockBroker(t, 1)
		metadataResponse := new(sarama.MetadataResponse)
		metadataResponse.AddBroker(broker.Addr(), broker.BrokerID())
		metadataResponse.AddTopicPartition("topic", 0, broker.BrokerID(), nil, nil, sarama.ErrNoError)
		broker.Returns(metadataResponse)

		KafkaInfo := NewKafkaInfo([]string{broker.Addr()}, "topic")

		Convey("When I make a producer", func() {

			p := KafkaInfo.makeProducer()

			Convey("Then the producer should be correctly set", func() {
				So(p, ShouldNotBeNil)
			})
		})
	})

	Convey("Given I create have a new kafka info with no kafka server listen", t, func() {

		KafkaInfo := NewKafkaInfo([]string{":1234"}, "topic")

		Convey("When I make a producer", func() {

			p := KafkaInfo.makeProducer()

			Convey("Then the producer should be nil", func() {
				So(p, ShouldBeNil)
			})
		})
	})
}

func TestKakfaInfo_makeConsumer(t *testing.T) {

	Convey("Given I create have a new kafka info with a kafka server listen", t, func() {

		broker := sarama.NewMockBroker(t, 1)
		metadataResponse := new(sarama.MetadataResponse)
		metadataResponse.AddBroker(broker.Addr(), broker.BrokerID())
		metadataResponse.AddTopicPartition("topic", 0, broker.BrokerID(), nil, nil, sarama.ErrNoError)
		broker.Returns(metadataResponse)

		KafkaInfo := NewKafkaInfo([]string{broker.Addr()}, "topic")

		Convey("When I make a consumer", func() {

			p := KafkaInfo.makeConsumer()

			Convey("Then the consumer should be correctly set", func() {
				So(p, ShouldNotBeNil)
			})
		})
	})

	Convey("Given I create have a new kafka info with no kafka server listen", t, func() {

		KafkaInfo := NewKafkaInfo([]string{":1234"}, "topic")

		Convey("When I make a producer", func() {

			p := KafkaInfo.makeConsumer()

			Convey("Then the consumer should be nil", func() {
				So(p, ShouldBeNil)
			})
		})
	})
}