Identity middleware
===================

Middleware component that authenticates requests against zebedee.

The identity and permissions returned from the identity endpoint are added to the request context.

### Getting started

Initialise the identity middleware and add it into the HTTP handler chain using `alice`:

```
    router := mux.NewRouter()
    alice := alice.New(identity.Handler(true)).Then(router)
    httpServer := server.New(config.BindAddr, alice)
```

Wrap authenticated endpoints using the `identity.Check(handler)` function to check that a request identity exists.

```
    router.Path("/jobs").Methods("POST").HandlerFunc(identity.Check(api.addJob))
```

Add required headers to outbound requests to other services

```
    import "github.com/ONSdigital/go-ns/common"

    common.AddServiceTokenHeader(req, api.AuthToken)
    common.AddUserHeader(req, "UserA")
```

or, put less portably:

```
    req.Header.Add("Authorization", api.AuthToken)
    req.Header.Add("User-Identity", "UserA")
```

But most of this should be done by `go-ns/rchttp` and `go-ns/clients/...`.

### Testing

If you need to use the middleware component in unit tests you can call the constructor function that allows injection of the HTTP client

```
import clientsidentity "github.com/ONSdigital/go-ns/clients/identity"
import "github.com/ONSdigital/go-ns/common/commontest"

httpClient := &commontest.RCHTTPClienterMock{
    DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
        return &http.Response{
            StatusCode: http.StatusOK,
        }, nil
    },
}
// set last argument to secretKey if you want to support legacy headers
clientsidentity.NewAPIClient(httpClient, zebedeeURL, "")

identityHandler := identity.HandlerForHTTPClient(doAuth, httpClient)
```
