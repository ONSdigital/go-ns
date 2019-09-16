go-ns [![Build Status](https://travis-ci.org/ONSdigital/go-ns.svg?branch=master)](https://travis-ci.org/ONSdigital/go-ns) [![GoDoc](https://godoc.org/github.com/ONSdigital/go-ns?status.svg)](https://godoc.org/github.com/ONSdigital/go-ns)
=====

###**_IMPORTANT NOTE_**
**This repository is to be considered _DEPRECATED_. No code should be added to this repository**

* The RHTTP and RCHTTP API clients have been removed from this repository. If in need of an HTTP API client then use the RCHTTP client in https://github.com/ONSdigital/dp-api-clients-go/

---

Common Go code for ONS apps:

* Common HTTP handlers for health check, requestID, and reverse proxy
* A logger which supports structured context-based logging
* Avro marshal and unmarshal functionality. Marshal function returns the avro encoding of an interface and the unmarshal function allows user to parse avro encoded byte array.
* [Kafka](./kafka/README.md) consumer and producer functionality

### Licence

Copyright ©‎ 2018, Crown Copyright (Office for National Statistics) (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
