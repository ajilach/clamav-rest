![Build Status](https://travis-ci.org/niilo/clamav-rest.svg) [![Docker Pulls](https://img.shields.io/docker/pulls/mashape/kong.svg)]()

This repository contains a basic rest api for clamav which allows sites to scan files as they are uploaded

!! Project needs cleenup, Makefile and documentation doesn't match current status !!

Usage:

build golang binary and docker image:
```bash
env GOOS=linux GOARCH=amd64 go build
docker build . -t niilo/clamav-rest
docker run -p 9000:9000 --rm -it niilo/clamav-rest
```

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