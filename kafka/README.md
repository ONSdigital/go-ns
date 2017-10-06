Kafka wrapper
=====

Use channels to abstract kafka consumers and producers.

For graceful handling of Closing consumers, it is advised to use the
`StopListeningToConsumer` method prior to the `Close`method. This will allow
inflight messages to be completed and successfully call commit so that the
message does not get replayed once the application restarts.

See the [example source file](../cmd/kafka-example/main.go) for a typical usage.
