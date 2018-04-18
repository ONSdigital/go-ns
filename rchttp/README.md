# rchttp

rchttp stands for robust contextual HTTP, and provides a default client
that inherits the methods associated with the standard HTTP client,
but with the addition of production-ready timeouts and context-sensitivity,
and the ability to perform exponential backoff when calling another HTTP server.

### How to use

rchttp should have a familiar feel to it when it is used - with an example given
below:

```go
import "github.com/ONSdigital/go-ns/rchttp"

func httpHandlerFunc(w http.ResponseWriter, req *http.Request) {
    client := rchttp.NewClient()

    resp, err := rcClient.Get(req.Context(), "https://www.google.com")
    if err != nil {
        fmt.Println(err)
        return
    }
}
```

In this case, in the unlikely event of https://www.google.com returning a status
of 500 or above, the client will retry at exponentially-increasing intervals, until
the max retries (10 by default is reached).

Also, if the inbound request is cancelled, for example, its context will be closed
and this will be noticed by the client.

You also do not have to use the default client if you don't like the configured
timeouts or do not wish to use exponential backoff. The following example shows
how to configure your own rchttp client:

```go
import "github.com/ONSdigital/go-ns/rchttp"

func main() {
    rcClient := &rchttp.Client{
        MaxRetries:         10,              // The maximum number of retries you wish to wait for
        ExponentialBackoff: true,            // Set to false if you do not want exponential backoff
        RetryTime:          1 * time.Second, // The time between the first set of retries

        HTTPClient: &http.Client{            // Create your own http client with configured timeouts
            Timeout: 10 * time.Second,
            Transport: &http.Transport{
                DialContext: (&net.Dialer{
                    Timeout: 5 * time.Second,
                }).DialContext,
                TLSHandshakeTimeout: 5 * time.Second,
                MaxIdleConns:        10,
                IdleConnTimeout:     30 * time.Second,
            },
        },
    }
}
```
