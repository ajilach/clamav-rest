[![Build Status](https://github.com/ajilach/clamav-rest/actions/workflows/ci.yaml/badge.svg)](https://github.com/ajilach/clamav-rest/actions/workflows/ci.yaml)
[![Latest Release](https://img.shields.io/github/v/release/ajilach/clamav-rest)](https://github.com/ajilach/clamav-rest/releases)
[![License: MIT](https://img.shields.io/github/license/ajilach/clamav-rest)](https://opensource.org/licenses/MIT)

# Table of Contents

- [Introduction](#introduction)
- [Installation](#installation)
- [Quick Start](#quick-start)
  - [Status Codes](#status-codes)
- [Endpoints](#endpoints)
  - [Utility endpoints](#utility-endpoints)
  - [Scanning endpoints](#scanning-endpoints)
- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [Networking](#networking)
- [Maintenance / Monitoring](#maintenance--monitoring)
  - [Shell Access](#shell-access)
  - [Prometheus](#prometheus)
- [Development](#development)
  - [Updates](#updates)
- [Deprecations](#deprecations)
  - [`/scan` Endpoint](#scan-endpoint)
    - [Differences between `/scan` and `/v2/scan`](#differences-between-scan-and-v2scan)
  - [centos.Dockerfile](#centosdockerfile)
- [Contributing](#contributing)
- [History](#history)
- [References](#references)
- [License](#license)

# Introduction

This is a two in one docker image which runs the open source virus scanner ClamAV (https://www.clamav.net/), performs automatic virus definition updates as a background process and provides a REST API interface to interact with the ClamAV process.

# Installation

Automated builds of the image are available on [Docker Hub](https://hub.docker.com/r/ajilaag/clamav-rest) and are the recommended method of installation. Grab the lastest release:

```bash
docker pull ajilaag/clamav-rest
```

The following image tags are available:

- `latest` - Most recent release of ClamAV with REST API
- `YYYYMMDD` - The day of the release
- `sha-...` - The git commit sha. This version ensures that the exact image is used and will be unique for each build

# Quick Start

> See [this docker-compose file](docker-compose-nonroot.yml) for non-root read-only usage.

Run clamav-rest docker image:

```bash
docker run -p 9000:9000 -p 9443:9443 -itd --name clamav-rest ajilaag/clamav-rest
```

The REST endpoints are now available on port 9000 (for http) and 9443 (for https).

If at least one virus is found, the API returns a `406 - Not Acceptable` response, a `200 - OK` otherwise.

Verify that the service detects common test virus signatures:

**HTTP:**

```bash
$ curl -i -F "file=@eicar.com.txt" http://localhost:9000/v2/scan
HTTP/1.1 100 Continue

HTTP/1.1 406 Not Acceptable
Content-Type: application/json; charset=utf-8
Date: Mon, 28 Aug 2017 20:22:34 GMT
Content-Length: 56

[{ "Status": "FOUND", "Description": "Eicar-Test-Signature","FileName":"eicar.com.txt"}]
```

**HTTPS:**

```bash
$ curl -i -k -F "file=@eicar.com.txt" https://localhost:9443/v2/scan
HTTP/1.1 100 Continue

HTTP/1.1 406 Not Acceptable
Content-Type: application/json; charset=utf-8
Date: Mon, 28 Aug 2017 20:22:34 GMT
Content-Length: 56

[{ "Status": "FOUND", "Description": "Eicar-Test-Signature","FileName":"eicar.com.txt"}]
```

Observe that the service returns `200` for a clean file:

**HTTP:**

```bash
$ curl -i -F "file=@clamrest.go" http://localhost:9000/v2/scan

HTTP/1.1 100 Continue

HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 28 Aug 2017 20:23:16 GMT
Content-Length: 33

[{ "Status": "OK", "Description": "","FileName":"clamrest.go"}]
```

**HTTPS:**

```bash
$ curl -i -k -F "file=@clamrest.go" https://localhost:9443/v2/scan

HTTP/1.1 100 Continue

HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 28 Aug 2017 20:23:16 GMT
Content-Length: 33

[{ "Status": "OK", "Description": "","FileName":"clamrest.go"}]
```

## Status Codes

- 200 - OK: clean file = no KNOWN infections
- 400 - ClamAV returned general error for file
- 406 - Not Acceptable: payload is infected
- 412 - Unable to parse the file provided
- 413 - Request entity too large: the file exceeds the scannable limit. Set MAX_FILE_SIZE to scan larger files
- 422 - Filename is missing in MimePart
- 501 - Unknown request

# Endpoints

## Utility endpoints

| Endpoint   | Description                                                                                                                 |
| ---------- | --------------------------------------------------------------------------------------------------------------------------- |
| `/`        | Home endpoint, returns stats for the currently running process                                                              |
| `/version` | Returns the clamav binary version and also the version of the virus signature databases and the signature last update date. |
| `/metrics` | Prometheus endpoint for scraping metrics.                                                                                   |

## Scanning endpoints

| Endpoint                 | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/v2/scan`               | Scanning endpoint, accepts a multipart/form-data request with one or more files and returns a json array with status, description and filename, along with the most severe http status code that was possible to determine. <br/><br/>**example response:** <br/> `[{"Status":"OK","Description":"","FileName":"checksums.txt"}]`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `/scanPath?path=/folder` | A scanning endpoint that will scan a folder. A practical example would be to mount a share into the container where you dump files into a folder, call `/scanPath` and let it scan the whole directory content, then continue processing them.<br/><br/>**example response:**<br/> `[{"Raw":"/folder: OK","Description":"","Path":"/folder","Hash":"","Size":0,"Status":"OK"}]`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `/scanHandlerBody`       | This endpoint scans the content in the HTTP POST request body.<br/><br/> **example response:**<br/> `{OK   200}`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `/scan`                  | [DEPRECATED] This endpoint scans in a similar manner to `/v2/scan` but does return one or more json objects without a valid json structure in between (no json array). It also does not include the filename as a json property. This endpoint is still present in the api for backwards compatibility for those who still use it, but it will also return headers indicating deprecation and pointing out the new, updated endpoint, `/v2/scan`. This endpoint does accept a multipart/form-data endpoint that by http standards can accept multiple files, and does scan them all, but the implementation of the endpoint indicates that it was originally (probably) meant to only scan one file at a time. Please don't rely on this endpoint to exist in the future. This project has the intention to sunset it to keep the project focus on a well maintainted set of features.<br/><br/>**example response:** <br/>`{"Status":"OK","Description":""}` |

# Configuration

## Environment Variables

Below is the complete list of available options that can be used to customize your installation.

| Parameter             | Description                                                                                             |
| --------------------- | ------------------------------------------------------------------------------------------------------- |
| `MAX_SCAN_SIZE`       | Amount of data scanned for each file. Defaults to `100M`                                                |
| `MAX_FILE_SIZE`       | Do not scan files larger than this size. Defaults to `25M`                                              |
| `MAX_RECURSION`       | How many nested archives to scan. Defaults to `16`                                                      |
| `MAX_FILES`           | Number of files to scan within an archive. Defaults to `10000`                                          |
| `MAX_EMBEDDEDPE`      | Maximum file size for embedded PE. Defaults to `10M`                                                    |
| `MAX_HTMLNORMALIZE`   | Maximum size of HTML to normalize. Defaults to `10M`                                                    |
| `MAX_HTMLNOTAGS`      | Maximum size of normlized HTML file to scan. Defaults to `2M`                                           |
| `MAX_SCRIPTNORMALIZE` | Maximum size of a script to normalize. Defaults to `5M`                                                 |
| `MAX_ZIPTYPERCG`      | Maximum size of ZIP to reanalyze type recognition. Defaults to `1M`                                     |
| `MAX_PARTITIONS`      | How many partitions per raw disk to scan. Defaults to `50`                                              |
| `MAX_ICONSPE`         | How many icons in PE to scan. Defaults to `100`                                                         |
| `PCRE_MATCHLIMIT`     | Maximum PCRE match calls. Defaults to `100000`                                                          |
| `PCRE_RECMATCHLIMIT`  | Maximum recursive match calls to PCRE. Defaults to `2000`                                               |
| `SIGNATURE_CHECKS`    | How many times per day to check for a new database signature. Must be between 1 and 50. Defaults to `2` |

## Networking

[TODO: is the description for port 3310 correct?]

| Port   | Description                              |
| ------ | ---------------------------------------- |
| `3310` | ClamD listening port (internal use only) |
| `9000` | HTTP REST listening port                 |
| `9443` | HTTPS REST listening port                |

# Maintenance / Monitoring

## Shell Access

For debugging and maintenance purposes you may want access the container's shell:

```bash
docker exec -it (whatever your container name is e.g. clamav-rest) /bin/sh
```

Checking the version with the `clamscan` command requires you to provide the custom database path.
The default value is `/clamav/data` set in the `/clamav/etc/clamd.conf` file, and the `clamav` service
was started with this `/clamav/etc/clamd.conf` referenced in `entrypoint.sh`.

```bash
clamscan --database=/clamav/data --version
```

## Prometheus

[Prometheus metrics](https://prometheus.io/docs/guides/go-application/) were implemented, which can be retrieved from the `/metrics` endpoint:

**HTTP:**  
`curl http://localhost:9000/metrics`

**HTTPS:**  
`curl https://localhost:9443/metrics`

Description of the metrics is available at these endpoints as part of the metrics themselves.

# Development

Source code can be found here: https://github.com/ajilach/clamav-rest

## Building the golang binary locally:

```sh
# For linux on amd64
GOOS=linux GOARCH=amd64 go build

# For macOS on arm64
GOOS=darwin GOARCH=arm64 go build

# For macOS on amd64
GOOS=darwin GOARCH=amd64 go build

# For Windows (using Git Bash or similar)
GOOS=windows GOARCH=amd64 go build
```

## Containerizing the application:

```bash
docker build . -t clamav-rest
docker run -p 9000:9000 -p 9443:9443 -itd --name clamav-rest clamav-rest
```

Note that the `docker build` command also takes care of compiling the source. Therefore you do not need to perform the manual build steps from above nor do you need a local go development environment.

## Updates

2025-02-07: Improved documentation.

2025-01-08: [PR 50](https://github.com/ajilach/clamav-rest/pull/50) integrated which now provides a new `/v2` endpoint returning more scan result information: status, description, http status and a list of scanned files. See the PR for more details. The old `/scan` endpoint is now considered deprecated. Also, a file size scan limit has been added which can be configured through the `MAX_FILE_SIZE` environment variable. This update also fixes a bug that would falsely return `200 OK` if the first file in a multi file scan was clean, regardless if any of the following files contained viruses. All endpoints now increment the Prometheus virus metric counter when a virus is discovered during a scan.

2024-10-21: freshclam notifies the correct `.clamd.conf` so that `clamd` is notified about updates and the correct version is returned now.
This is an additional fix to the latest fix from October 15 2024 which was not working. Thanks to [christianbumann](https://github.com/christianbumann) and [arizon-dread](https://github.com/arizon-dread).

2024-10-15: ClamAV was thought to handle database updates correctly thanks to [christianbumann](https://github.com/christianbumann). It turned out that this was not the case.

As of May 2024, the releases are built for multiple architectures thanks to efforts from [kcirtapfromspace](https://github.com/kcirtapfromspace) and support non-root read-only deployments thanks to [robaca](https://github.com/robaca).

The additional endpoint `/version` is now available to check the `clamd` version and signature date. Thanks [pastral](https://github.com/pastral).

Closed a security hole by upgrading our `Dockerfile` to the alpine base image version `3.19` thanks to [Marsup](https://github.com/Marsup).

# Deprecations

## `/scan` endpoint

As of release [20250109](https://github.com/ajilach/clamav-rest/releases/tag/20250109) the `/scan` endpoint is deprecated and `/v2/scan` is now the preferred endpoint to use.

### Differences between `/scan` and `/v2/scan`

Since the endpoint can receive one or several files, the response has been updated to always be returned as a json array and the filename is now included as a property in the response, to make it easy to find out what file(s) contains virus.

## Centos Dockerfile

The [centos.Dockerfile](./centos.Dockerfile) has been last updated in the release [20250109](https://github.com/ajilach/clamav-rest/releases/tag/20250109) but will not be maintained anymore going forward. If there are community members using it, please consider contributing.

# Contributing

We welcome and appreciate contributions from the community. To keep our project maintainable and high quality, please follow these best practices:

- **Fork and Branch:** Fork the repository and work on a feature branch. Make sure your branch is up-to-date with the latest changes.
- **Coding Standards:** Adhere to standard Go conventions and ensure your code is clean, well-documented, and tested.
- **Commit Messages:** Write clear and concise commit messages explaining your changes.
- **Pull Requests:** Open a pull request with a clear description of your changes and reference any related issues. Our maintainers will review and provide feedback.
- **Issues:** If you encounter a bug or have a feature suggestion, please open an issue before starting work to discuss your idea.
- **Documentation:** Update relevant documentation and tests as needed with your changes.

Thank you for helping improve the project!

# History

This work is based on the awesome work done by [o20ne/clamav-rest](https://github.com/o20ne/clamav-rest) which is based on [niilo/clamav-rest](https://github.com/niilo/clamav-rest) which in turn is based on the original code from [osterzel/clamav-rest](https://github.com/osterzel/clamav-rest).

# References

- [The ClamAV project](https://www.clamav.net)
- [The ajilach/clamav-rest project](https://github.com/ajilach/clamav-rest)

# License

This project is licensed under the MIT License. See the [LICENSE](LICENSE.md) file for details.
