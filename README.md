# Table of Contents

- [Introduction](#introduction)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
    - [Status Codes](#status-codes)
- [Configuration](#configuration)
    - [Environment Variables](#environment-variables)
    - [Networking](#networking)
- [Maintenance / Monitoring](#maintenance--monitoring)
    - [Shell Access](#shell-access)

- [Developing](#developing)
- [References](#references)

# Introduction

This is two in one docker image so it runs open source virus scanner ClamAV (https://www.clamav.net/), automatic virus definition updates as background process and REST API interface to interact with ClamAV process.

# Prerequisites

This container doesn't do much on it's own unless you use an additional service or communicator to talk to it!

# Installation

Automated builds of the image are available on [Registry](https://hub.docker.com/r/ajilaag/clamav-rest) and is the recommended method of installation.

```bash
docker pull hub.docker.com/ajilaag/clamav-rest:(imagetag)
```

The following image tags are available:
* `latest` - Most recent release of ClamAV with REST API

# Quick Start

Run clamav-rest docker image:
```bash
docker run -p 9000:9000 -p 9443:9443 -itd --name clamav-rest ajilaag/clamav-rest
```

Test that service detects common test virus signature:

**HTTP**
```bash
$ curl -i -F "file=@eicar.com.txt" http://localhost:9000/scan
HTTP/1.1 100 Continue

HTTP/1.1 406 Not Acceptable
Content-Type: application/json; charset=utf-8
Date: Mon, 28 Aug 2017 20:22:34 GMT
Content-Length: 56

{ "Status": "FOUND", "Description": "Eicar-Test-Signature" }
```

**HTTPS**
```bash
$ curl -i -k -F "file=@eicar.com.txt" https://localhost:9443/scan
HTTP/1.1 100 Continue

HTTP/1.1 406 Not Acceptable
Content-Type: application/json; charset=utf-8
Date: Mon, 28 Aug 2017 20:22:34 GMT
Content-Length: 56

{ "Status": "FOUND", "Description": "Eicar-Test-Signature" }
```

Test that service returns 200 for clean file:

**HTTP**
```bash
$ curl -i -F "file=@clamrest.go" http://localhost:9000/scan

HTTP/1.1 100 Continue

HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 28 Aug 2017 20:23:16 GMT
Content-Length: 33

{ "Status": "OK", "Description": "" }
```
**HTTPS**
```bash
$ curl -i -k -F "file=@clamrest.go" https://localhost:9443/scan

HTTP/1.1 100 Continue

HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 28 Aug 2017 20:23:16 GMT
Content-Length: 33

{ "Status": "OK", "Description": "" }
```



## Status Codes
- 200 - clean file = no KNOWN infections
- 400 - ClamAV returned general error for file
- 406 - INFECTED
- 412 - unable to parse file
- 501 - unknown request

# Configuration

## Environment Variables

Below is the complete list of available options that can be used to customize your installation.

| Parameter | Description |
|-----------|-------------|
| `MAX_SCAN_SIZE` | Amount of data scanned for each file - Default `100M` |
| `MAX_FILE_SIZE` | Don't scan files larger than this size - Default `25M` |
| `MAX_RECURSION` | How many nested archives to scan - Default `16` |
| `MAX_FILES` | Number of files to scan withn archive - Default `10000` |
| `MAX_EMBEDDEDPE` | Maximum file size for embedded PE - Default `10M` |
| `MAX_HTMLNORMALIZE` | Maximum size of HTML to normalize - Default `10M` |
| `MAX_HTMLNOTAGS` | Maximum size of Normlized HTML File to scan- Default `2M` |
| `MAX_SCRIPTNORMALIZE` | Maximum size of a Script to normalize - Default `5M` |
| `MAX_ZIPTYPERCG` | Maximum size of ZIP to reanalyze type recognition - Default `1M` |
| `MAX_PARTITIONS` | How many partitions per Raw disk to scan - Default `50` |
| `MAX_ICONSPE` | How many Icons in PE to scan - Default `100` |
| `PCRE_MATCHLIMIT` | Maximum PCRE Match Calls - Default `100000` |
| `PCRE_RECMATCHLIMIT` | Maximum Recursive Match Calls to PCRE - Default `2000` |
| `SIGNATURE_CHECKS` | Check times per day for a new database signature. Must be between 1 and 50. - Default `2` |
| `PROXY_SERVER` | Specify a proxy for freshclam to utilize, if applicable, set in environment variables - Optional |
| `PROXY_PORT` | The port for the proxy server, if applicable, set in environment variables - Optional |
| `PROXY_USERNAME` | The username for the proxy server, if applicable, set in environment variables - Optional |
| `PROXY_PASSWORD` | The password for the proxy server, if applicable, set in environment variables - Optional |

## Networking

| Port | Description |
|-----------|-------------|
| `3310`    | ClamD Listening Port |

# Maintenance / Monitoring

## Shell Access

For debugging and maintenance purposes you may want access the containers shell.

```bash
docker exec -it (whatever your container name is e.g. clamav-rest) /bin/sh
```
## Prometheus

[Prometheus metrics](https://prometheus.io/docs/guides/go-application/) were implemented, which can be retrieved as follows

**HTTP**:
curl http://localhost:9000/metrics

**HTTPS:**
curl https://localhost:9443/metrics

# Developing

Source Code can be found here: https://github.com/ajilach/clamav-rest

Build golang (linux) binary and docker image:

```bash
# env GOOS=linux GOARCH=amd64 go build
docker build . -t clamav-go-rest
docker run -p 9000:9000 -p 9443:9443 -itd --name clamav-rest clamav-go-rest
```

# References

* https://www.clamav.net
* https://github.com/ajilach/clamav-rest
