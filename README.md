Build status:
- [![Build Status](https://travis-ci.org/osterzel/clamav-rest.svg)]

This repository contains a basic rest api for clamav which allows sites to scan files as they are uploaded

Usage:

make build-container - this will build the docker container and link it up to a clamav container
                       it exposes a rest api on port 9000

make test - this will test against the rest api
