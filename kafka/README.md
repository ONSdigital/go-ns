Kafka wrapper
=====

Use channels to abstract kafka consumers and producers.

For graceful handling of Closing consumers, it is advised to use the `StopListeningToConsumer` method prior to the `Close` method. This will allow inflight messages to be completed and successfully call commit so that the message does not get replayed once the application restarts.

It is recommended to use `NewSyncConsumer` - where, when you have read a message from `Incoming()`,
the listener for messages will block (and not read the next message from kafka)
until you signal that the message has been consumed (typically with `CommitAndRelease(msg)`).
Otherwise, if the application gets shutdown (e.g. interrupt signal), and has to be shutdown,
the consumer may not be shutdown in a timely manner (because it is blocked sending the read message to `Incoming()`).

See the [example source file](../cmd/kafka-example/main.go) for a typical usage.
