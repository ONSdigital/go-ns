package kafka_test

import (
	"context"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/kafka/kafkatest"
	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestAsyncConsumer_Close(t *testing.T) {

	Convey("Given a consumer", t, func() {

		messages := make(chan kafka.Message, 1)
		mockConsumer := kafkatest.NewMessageConsumer(messages)

		message := kafkatest.NewMessage([]byte("message contents"))
		messages <- message

		consumer := kafka.NewAsyncConsumer()

		consumer.Consume(mockConsumer, func(message kafka.Message) {
			log.Info("consuming message", nil)
		})

		Convey("When close is called", func() {

			err := consumer.Close(context.Background())

			Convey("Then no errors are returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestAsyncConsumer_Consume(t *testing.T) {
	Convey("Given a consumer", t, func() {

		messages := make(chan kafka.Message, 1)
		mockConsumer := kafkatest.NewMessageConsumer(messages)

		expectedMessageContents := "message contents"
		message := kafkatest.NewMessage([]byte(expectedMessageContents))
		messages <- message

		consumer := kafka.NewAsyncConsumer()

		var actualMessageContent string
		handlerFunc := func(message kafka.Message) {
			// for testing just take the message content and assign it to the var outside the function.
			// allowing us to check its contents
			actualMessageContent = string(message.GetData())
		}

		Convey("When consume is called with a test handler function", func() {

			consumer.Consume(mockConsumer, handlerFunc)

			waitForMessageToBeConsumed(&actualMessageContent)

			Convey("Then the handler function passed to consume is called with the test message", func() {
				So(actualMessageContent, ShouldEqual, expectedMessageContents)
			})
		})
	})
}

func waitForMessageToBeConsumed(messageContent *string) {

	start := time.Now()
	timeout := start.Add(time.Millisecond * 500)
	for {
		if len(*messageContent) > 0 {
			log.Debug("message has been consumed", nil)
			break
		}

		if time.Now().After(timeout) {
			log.Debug("timeout hit", nil)
			break
		}

		log.Debug("not yet consumed, waiting a bit longer", nil)
		time.Sleep(time.Millisecond * 10)
	}
}
