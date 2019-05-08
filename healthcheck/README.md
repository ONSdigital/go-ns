Health checks
=============

#### Add a health check to a service without an existing HTTP server

Create a health check server to check neo4j and the filter API every 30 seconds:
```
neoHealthChecker := neo4j.NewHealthCheckClient(neo4jConnPool)
elasticsearchChecker := elasticsearch.NewHealthCheckClient(url)
filterAPIHealthChecker := filterHealthCheck.New(config.FilterAPIURL)

healthChecker := healthcheck.NewServer(
    config.BindAddr,
    config.HealthCheckInterval,
    errorChannel,
    filterAPIHealthChecker,
    elasticsearchChecker,
    neoHealthChecker,
)
```

Make sure you call close on shutdown:

```
err = healthChecker.Close(ctx)
```

#### Add a health check to a service with an existing HTTP server

Register the health check handler as a route:
```
router.Path("/healthcheck").HandlerFunc(healthcheck.Do)
```

Create a healthcheck.Ticker to call the given client at regular intervals
```
ticker := healthcheck.NewTicker(duration, clients...)
```

Make sure you call ticker.Close() on shutdown to release resources:

```
ticker.Close()
```

#### Existing health check clients

The `clients` package in `go-ns` provides clients for the DP API's that can be used as health check clients.

There are also `healthcheck.Client` implementations for other services in go-ns packages for those services.

#### Creating new health check clients

A default healthcheck client (that uses rchttp.DefaultClient to call the service endpoint) can be obtained by calling
```
healthcheck.NewDefaultClient(service, url)
```
e.g.
```
client := healthcheck.NewDefaultClient("service name", "http://service-host:80/healthcheck")
```

If you don't want to use the default client,
any implementation of the healthcheck.Client interface can be used as a client:
```
type Client interface {
	Healthcheck() (string, error)
}
```

The function should return the service name, and an error if one occurred on the health check.
