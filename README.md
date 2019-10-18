go-ns [![Build Status](https://travis-ci.org/ONSdigital/go-ns.svg?branch=master)](https://travis-ci.org/ONSdigital/go-ns) [![GoDoc](https://godoc.org/github.com/ONSdigital/go-ns?status.svg)](https://godoc.org/github.com/ONSdigital/go-ns)
=====

### **_DEPRECATED_**

**This repository is to be considered _DEPRECATED_. No code should be added to this repository**

The following library changes have been made as part of deprecating this repo:

* The `rhttp` client has been removed and should no longer be used.  Any usage of `rhttp` should be replaced by `rchttp`.
* The `rchttp` client has been moved to a new repo [dp-rchttp](https://github.com/ONSdigital/dp-rchttp).  All imports should be updated accordingly.
* The API clients have been moved to a new repo [dp-api-clients-go](https://github.com/ONSdigital/dp-api-clients-go/).  All imports should be updated accordingly.

---

Common Go code for ONS apps:

* Common HTTP handlers for health check, requestID, and reverse proxy
* A logger which supports structured context-based logging
* Avro marshal and unmarshal functionality. Marshal function returns the avro encoding of an interface and the unmarshal function allows user to parse avro encoded byte array.
* [Kafka](./kafka/README.md) consumer and producer functionality

### Licence

Copyright ©‎ 2018, Crown Copyright (Office for National Statistics) (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
