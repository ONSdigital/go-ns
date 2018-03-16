Identity middleware
===================

Middleware componenet that authenticates requests against zebedee.

The identity and permissions returned from the identity endpoint are added to the request context.

### Getting started

Initialise the identity middleware and add it into the HTTP handler chain using alice:

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
    req.Header.Add("Authorization", api.AuthToken)
    req.Header.Add("User-Identity", ctx.Value("User-Identity").(string))
```

### Testing

If you need to use the middleware component in unit tests you can call the constructor function that allows injection of the HTTP client

```
httpClient := &identitytest.HttpClientMock{
    DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
        return &http.Response{
            StatusCode: http.StatusOK,
        }, nil
    },
}

identityHandler := identity.HandlerForHttpClient(doAuth, httpClient)
```