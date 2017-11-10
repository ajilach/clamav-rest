![Build Status](https://travis-ci.org/niilo/clamav-rest.svg) [![Docker Pulls](https://img.shields.io/docker/pulls/niilo/clamav-rest.svg)]()

This is two in one docker image so it runs open source virus scanner ClamAV (https://www.clamav.net/), automatic virus definition updates as background process and REST api interface to interact with ClamAV process.

Travis CI build will build new release on weekly basis and push those to Docker hub [ClamAV-rest docker image](https://hub.docker.com/r/niilo/clamav-rest/). Virus definitions will be updated on every docker build.


## Usage:

Run clamav-rest docker image:
```bash
docker run -p 9000:9000 --rm -it niilo/clamav-rest
```

Test that service detects common test virus signature:
```bash
$ curl -i -F "file=@eicar.com.txt" http://localhost:9000/scan
HTTP/1.1 100 Continue

HTTP/1.1 406 Not Acceptable
Content-Type: application/json; charset=utf-8
Date: Mon, 28 Aug 2017 20:22:34 GMT
Content-Length: 56

{ Status: "FOUND", Description: "Eicar-Test-Signature" }
```

Test that service returns 200 for clean file:
```bash
$ curl -i -F "file=@clamrest.go" http://localhost:9000/scan

HTTP/1.1 100 Continue

HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 28 Aug 2017 20:23:16 GMT
Content-Length: 33

{ Status: "OK", Description: "" }
```

**Status codes:**
- 200 - clean file = no KNOWN infections
- 406 - INFECTED
- 400 - ClamAV returned general error for file
- 412 - unable to parse file
- 501 - unknown request


## Developing:

Build golang (linux) binary and docker image:
```bash
env GOOS=linux GOARCH=amd64 go build
docker build . -t niilo/clamav-rest
docker run -p 9000:9000 --rm -it niilo/clamav-rest
```
