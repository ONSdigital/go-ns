### Testing auditing

The `audit_mocks` package provides useful methods for verifying calls to `audit.Record()` reducing the amount of 
code duplication to setup a mock auditor and verify its invocations during a test case.

### Getting started
Create an auditor mock that returns no error.
```go
auditor := audit_mock.New()
```
Create an auditor mock that returns an error when `Record()` is called with particular action and result values
```go
auditor := audit_mock.NewErroring("some task", "the outcome")
```
Assert `auditor.Record()` is called the expected number of times and the `action`, `result` and `auditParam` values in
 each call are as expected.
```go
auditor.AssertRecordCalls(
    audit_mock.ExpectedParams{"my_action", audit.Attempted, common.Params{"key":"value"},
    audit_mock.ExpectedParams{instance.GetInstancesAction, audit.Successful, nil},
)
```