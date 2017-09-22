package errorhandler_test

import (
	"errors"
	"testing"

	"github.com/ONSdigital/go-ns/errorhandler/mock"

	"github.com/ONSdigital/go-ns/errorhandler"
	"github.com/ONSdigital/go-ns/kafka/kafkatest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRun(t *testing.T) {
	Convey("Given an error handler", t, func() {
		Convey("When the error handler is called", func() {
			errHandle := mocks.HandlerMock{
				HandleFunc: func(instanceId string, err error) {
					//
				},
			}
			errHandle.Handle("instanceId", errors.New("text"))
			Convey("A complete run through should conists of 1 call to the handler", func() {
				So(len(errHandle.HandleCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestKafkaProducer(t *testing.T) {
	Convey("Given a new error kafka producer is created", t, func() {
		Convey("When a new kafka producer is created ", func() {
			outputChannel := make(chan []byte, 1)
			mockMessageProducer := kafkatest.NewMessageProducer(outputChannel, nil, nil)
			errHandle := errorhandler.NewKafkaHandler(mockMessageProducer)
			Convey("And the error kakfa is not nil", func() {
				So(errHandle, ShouldNotBeNil)
			})
		})
	})
}
