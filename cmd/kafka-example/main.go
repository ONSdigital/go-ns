package main

import (
	"bufio"
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Brokers       []string      `envconfig:"KAFKA_ADDR"`
	KafkaMaxBytes int           `envconfig:"KAFKA_MAX_BYTES"`
	ConsumedTopic string        `envconfig:"KAFKA_CONSUMED_TOPIC"`
	ConsumedGroup string        `envconfig:"KAFKA_CONSUMED_GROUP"`
	ProducedTopic string        `envconfig:"KAFKA_PRODUCED_TOPIC"`
	ConsumeMax    int           `envconfig:"KAFKA_CONSUME_MAX"`
	KafkaSync     bool          `envconfig:"KAFKA_SYNC"`
	TimeOut       time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	Chomp         bool          `envconfig:"CHOMP_MSG"`
	Snooze        bool          `envconfig:"SNOOZE"`
	OverSleep     bool          `envconfig:"OVERSLEEP"`
}

func main() {
	log.Namespace = "kafka-example"
	cfg := &Config{
		Brokers:       []string{"locahost:9092"},
		KafkaMaxBytes: 50 * 1024 * 1024,
		KafkaSync:     true,
		ConsumedGroup: log.Namespace,
		ConsumedTopic: "input",
		ProducedTopic: "output",
		ConsumeMax:    0,
		TimeOut:       5 * time.Second,
		Chomp:         false,
		Snooze:        false,
		OverSleep:     false,
	}
	if err := envconfig.Process("", cfg); err != nil {
		panic(err)
	}

	log.Info("[KAFKA-TEST] Starting (consumer sent to stdout, stdin sent to producer)",
		log.Data{"consumed_group": cfg.ConsumedGroup, "consumed_topic": cfg.ConsumedTopic, "produced_topic": cfg.ProducedTopic})

	kafka.SetMaxMessageSize(int32(cfg.KafkaMaxBytes))
	producer, err := kafka.NewProducer(cfg.Brokers, cfg.ProducedTopic, cfg.KafkaMaxBytes)
	if err != nil {
		log.ErrorC("[KAFKA-TEST] Could not create producer", err, nil)
		panic("[KAFKA-TEST] Could not create producer")
	}

	var consumer *kafka.ConsumerGroup
	if cfg.KafkaSync {
		consumer, err = kafka.NewSyncConsumer(cfg.Brokers, cfg.ConsumedTopic, cfg.ConsumedGroup, kafka.OffsetNewest)
	} else {
		consumer, err = kafka.NewConsumerGroup(cfg.Brokers, cfg.ConsumedTopic, cfg.ConsumedGroup, kafka.OffsetNewest)
	}
	if err != nil {
		log.ErrorC("[KAFKA-TEST] Could not create consumer", err, nil)
		panic("[KAFKA-TEST] Could not create consumer")
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	stdinChannel := make(chan string)

	eventLoopContext, eventLoopCancel := context.WithCancel(context.Background())
	eventLoopDone := make(chan bool)
	consumeCount := 0

	go func(ch chan string) {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if cfg.Chomp {
				line = line[:len(line)-1]
			}
			ch <- line
		}
		eventLoopCancel()
		<-eventLoopDone
		close(ch)
	}(stdinChannel)

	// eventLoop
	go func() {
		defer close(eventLoopDone)
		for {
			select {
			case <-time.After(1 * time.Second):
				log.Trace("[KAFKA-TEST] tick", nil)
			case <-eventLoopContext.Done():
				log.Trace("[KAFKA-TEST] Event loop context done", log.Data{"eventLoopContextErr": eventLoopContext.Err()})
				return
			case consumedMessage := <-consumer.Incoming():
				consumeCount++
				logData := log.Data{"consumeCount": consumeCount, "consumeMax": cfg.ConsumeMax, "messageOffset": consumedMessage.Offset()}
				log.Info("[KAFKA-TEST] Received message", logData)

				consumedData := consumedMessage.GetData()
				logData["messageString"] = string(consumedData)
				logData["messageRaw"] = consumedData
				logData["messageLen"] = len(consumedData)

				var sleep time.Duration
				if cfg.Snooze || cfg.OverSleep {
					// Snooze slows consumption for testing
					sleep = 500 * time.Millisecond
					if cfg.OverSleep {
						// OverSleep tests taking more than shutdown timeout to process a message
						sleep += cfg.TimeOut
					}
					logData["sleep"] = sleep
				}

				log.Info("[KAFKA-TEST] Message consumed", logData)
				if sleep > time.Duration(0) {
					time.Sleep(sleep)
					log.Trace("[KAFKA-TEST] done sleeping", nil)
				}

				// send downstream
				producer.Output() <- consumedData

				if cfg.KafkaSync {
					log.Trace("[KAFKA-TEST] pre-release", nil)
					consumer.CommitAndRelease(consumedMessage)
				} else {
					log.Trace("[KAFKA-TEST] pre-commit", nil)
					consumedMessage.Commit()
				}
				log.Info("[KAFKA-TEST] committed message", log.Data{"messageOffset": consumedMessage.Offset()})
				if consumeCount == cfg.ConsumeMax {
					log.Trace("[KAFKA-TEST] consumed max - exiting eventLoop", nil)
					return
				}
			case stdinLine := <-stdinChannel:
				producer.Output() <- []byte(stdinLine)
				log.Info("[KAFKA-TEST] Message output", log.Data{"messageSent": stdinLine, "messageChars": []byte(stdinLine)})
			}
		}
	}()

	// block until a fatal error, signal or eventLoopDone - then proceed to shutdown
	select {
	case <-eventLoopDone:
		log.Info("[KAFKA-TEST] Quitting after event loop aborted", nil)
	case sig := <-signals:
		log.Info("[KAFKA-TEST] Quitting after OS signal", log.Data{"signal": sig})
	case consumerError := <-consumer.Errors():
		log.ErrorC("[KAFKA-TEST] Aborting consumer", consumerError, nil)
	case producerError := <-producer.Errors():
		log.ErrorC("[KAFKA-TEST] Aborting producer", producerError, nil)
	}

	// give the app `Timeout` seconds to close gracefully before killing it.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.TimeOut)

	// background graceful shutdown
	go func() {
		log.Info("[KAFKA-TEST] Stopping kafka consumer listener", nil)
		consumer.StopListeningToConsumer(ctx)
		log.Info("[KAFKA-TEST] Stopped kafka consumer listener", nil)
		eventLoopCancel()
		// wait for eventLoopDone: all in-flight messages have been processed
		<-eventLoopDone
		log.Info("[KAFKA-TEST] Closing kafka producer", nil)
		producer.Close(ctx)
		log.Info("[KAFKA-TEST] Closed kafka producer", nil)
		log.Info("[KAFKA-TEST] Closing kafka consumer", nil)
		consumer.Close(ctx)
		log.Info("[KAFKA-TEST] Closed kafka consumer", nil)

		log.Info("[KAFKA-TEST] Done shutdown - cancelling timeout context", nil)
		cancel() // stop timer
	}()

	// wait for timeout or success (via cancel)
	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		log.ErrorC("[KAFKA-TEST]", ctx.Err(), nil)
	} else {
		log.Info("[KAFKA-TEST] Done shutdown gracefully", log.Data{"context": ctx.Err()})
	}
	os.Exit(1)
}
