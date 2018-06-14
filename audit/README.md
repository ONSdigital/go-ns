## Auditing

### Creating an auditor

To create a new auditor simply provide a Kafka producer (see go-ns/kafka) for the topic you wish to send your audit 
events to and the name of the service auditing the even. 
```go
auditor = audit.New(auditProducer, "dp-dataset-api")
```

You can also create a nop auditor which satisfies the `Auditor` interface but does nothing when called.

```go
auditor = &audit.NopAuditor{}
```

### Recording events
To record an event simply call `Auditor.Record()` passing in the appropriate arguments for the event you wish to record.
The following example is a typical use case for recording an audit event.
```go
// audit params is map holding additional useful information for the event
auditParams := common.Params{"key":"value"}

// audit that an action has been attempted
if err := auditor.Record(ctx, "my_action", audit.Attempted, auditParams); err != nil {
    // handle error
}

// attempt do carry out the action
err := func() error {
    // business logic...
}()

if err != nil {
    // attempted action unsuccessful - record unsuccessful event
    if err := auditor.Record(ctx, "my_action", audit.Unsuccessful, auditParams); err != nil {
        // handle error
    } 
    // handle error
}

// action completed successfully - record success event
if err := auditor.Record(ctx, "my_action", audit.Successful, auditParams); err != nil {
    // handle error
} 
```
`Auditor.Record()` will automatically extract `requestID`/`correlationID`, `User-Identity` & `Caller-Identity` from the
 supplied context (if they exist) and add them to the audit event and log parameters.
 
If `Auditor.Record()` fails to record the event it will log the error (including `requestID`/`correlationID`, 
`User-Identity` & `Caller-Identity` if they are available) before returning.