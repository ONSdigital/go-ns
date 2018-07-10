### Testing auditing

The `auditortest` package provides useful methods for verifying calls to `audit.Record()` reducing the amount of
code duplication to setup a mock auditor and verify its invocations during a test case.

### Getting started
Create an auditor mock that returns no error.
```go
auditor :=  auditortest.New(t)
```
Create an auditor mock that returns an error when `Record()` is called with particular action and result values
```go
auditor := auditortest.NewErroring(t, "some task", "the outcome")
```
Assert `auditor.Record()` is called the expected number of times and the `action`, `result` and `auditParam` values in
 each call are as expected.
```go
auditor.AssertRecordCalls(
    auditortest.Expected{"my_action", audit.Attempted, common.Params{"key":"value"}},
    auditortest.Expected{instance.GetInstancesAction, audit.Successful, nil},
)
```
