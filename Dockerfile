FROM golang:alpine

# Update
RUN apk update upgrade;

# Set timezone to Europe/Zurich
RUN apk add tzdata
RUN ln -s /usr/share/zoneinfo/Europe/Zurich /etc/localtime

# Install ClamAV
RUN apk --no-cache add clamav clamav-libunrar \
    && mkdir /run/clamav \
    && chown clamav:clamav /run/clamav

# Configure clamAV to run in foreground with port 3310
RUN sed -i 's/^#Foreground .*$/Foreground true/g' /etc/clamav/clamd.conf \
    && sed -i 's/^#TCPSocket .*$/TCPSocket 3310/g' /etc/clamav/clamd.conf \
    && sed -i 's/^#Foreground .*$/Foreground true/g' /etc/clamav/freshclam.conf

RUN freshclam --quiet --no-dns --checks=2

# Build go package
ADD . /go/src/clamav-rest/
RUN cd /go/src/clamav-rest && go build -v

COPY entrypoint.sh /usr/bin/
RUN mv /go/src/clamav-rest/clamav-rest /usr/bin/ && rm -Rf /go/src/clamav-rest

EXPOSE 9000

ENTRYPOINT [ "entrypoint.sh" ]