go-ns [![GoDoc](https://godoc.org/github.com/ONSdigital/go-ns?status.svg)](https://godoc.org/github.com/ONSdigital/go-ns)
=====

### **_DEPRECATED_**

**This repository is to be considered _DEPRECATED_. No code should be added to this repository**

The following library changes have been made as part of deprecating this repo:

* The `rhttp` client has been removed and should no longer be used.  Any usage of `rhttp` should be replaced by [dp-net/http](https://github.com/ONSdigital/dp-net).
* The `rchttp` client has been moved to a new repo [dp-net/http](https://github.com/ONSdigital/dp-net).  All imports should be updated accordingly.
* The API clients have been moved to a new repo [dp-api-clients-go](https://github.com/ONSdigital/dp-api-clients-go/).  All imports should be updated accordingly.
* The `healthcheck` library has been moved to a new repo [dp-healthcheck](https://github.com/ONSdigital/dp-healthcheck) and updated inline with the [latest spec](https://github.com/ONSdigital/dp/blob/master/standards/HEALTH_CHECK_SPECIFICATION.md). All apps should be updated to use the new health check implementation.
* The `healthcheck` client has been reimplemented inline with the [latest spec](https://github.com/ONSdigital/dp/blob/master/standards/HEALTH_CHECK_SPECIFICATION.md) in [dp-api-clients-go/health](https://github.com/ONSdigital/dp-api-clients-go/tree/master/health). All clients in `dp-api-clients-go` have been updated to use this new library so all apps should be updated to use these new clients.
* The `healthcheck` handler has been reimplemented inline with the [latest spec](https://github.com/ONSdigital/dp/blob/master/standards/HEALTH_CHECK_SPECIFICATION.md) in [dp-healthcheck](https://github.com/ONSdigital/dp-healthcheck). All `/healthcheck` routes should be changed to `/health` and updated to use this new handler.
* The `elasticsearch` client has been moved to [dp-elasticsearch](https://github.com/ONSdigital/dp-elasticsearch). All imports should be updated accordingly.
* The `kafka` client has been moved to [dp-kafka](https://github.com/ONSdigital/dp-kafka). All imports should be updated accordingly.
* The `mongo` client has been moved to [dp-mongodb](https://github.com/ONSdigital/dp-mongodb). All imports should be updated accordingly.
* The `neo4j` client has been moved to [dp-graph](https://github.com/ONSdigital/dp-graph). All usages of the old `neo4j` library should be replaced with `dp-graph` using the `neo4jdriver`.
* The `s3` client has been moved to [dp-s3](https://github.com/ONSdigital/dp-s3). All imports should be updated accordingly.
* The `vault` client has been moved to [dp-vault](https://github.com/ONSdigital/dp-vault). All imports should be updated accordingly.
* The `log` library has been replaced by [log.go](https://github.com/ONSdigital/log.go) which has been implemented in line with the new [logging standards](https://github.com/ONSdigital/dp/blob/master/standards/LOGGING_STANDARDS.md). A [script is provided](https://github.com/ONSdigital/log.go/blob/master/scripts/edit-logs.sh) to automate the replacement of the log library in most common usages.
* The `common` library has been moved to [dp-net/request](https://github.com/ONSdigital/dp-net). All imports should be updated accordingly.
* The `server` library has been moved to [dp-net/http](https://github.com/ONSdigital/dp-net). All imports should be updated accordingly.
* The `request` library has been moved to [dp-net/http](https://github.com/ONSdigital/dp-net). All imports should be updated accordingly.
* The `handlers` library has been moved to [dp-net/handlers](https://github.com/ONSdigital/dp-net). All imports should be updated accordingly.
* The `identity` library has been moved to [dp-net/handlers](https://github.com/ONSdigital/dp-net). All imports should be updated accordingly.

---

Common Go code for ONS apps:

* Common HTTP handlers for health check, requestID, and reverse proxy
* A logger which supports structured context-based logging
* Avro marshal and unmarshal functionality. Marshal function returns the avro encoding of an interface and the unmarshal function allows user to parse avro encoded byte array.
* [Kafka](./kafka/README.md) consumer and producer functionality

### Licence

Copyright ©‎ 2018, Crown Copyright (Office for National Statistics) (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
