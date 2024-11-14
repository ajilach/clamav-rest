FROM quay.io/centos/centos:stream9 as build

# Set timezone to Europe/Zurich
ENV TZ=Europe/Zurich

# Install golang
RUN mkdir -p /go && chmod -R 777 /go \
    && dnf -y update \
    && dnf -y group install "Development Tools" \
    && dnf install -y epel-release \
    && dnf -y install golang \
    && dnf clean all

ENV GOPATH=/go \
    PATH="$GOPATH/bin:/usr/local/go/bin:$PATH"

# Build go package
ADD . /go/src/clamav-rest/
RUN cd /go/src/clamav-rest && go mod tidy && go build -v

FROM quay.io/centos/centos:stream9

# Copy compiled clamav-rest binary from build container to production container
COPY --from=build /go/src/clamav-rest/clamav-rest /usr/bin/

# Install ClamAV
RUN dnf -y update \
    && dnf install -y epel-release \
    && dnf install -y clamav-server clamav-data clamav-update clamav-filesystem clamav clamav-scanner-systemd clamav-devel clamav-lib clamav-server-systemd \
    && mkdir /run/clamav \
    && chown clamscan:clamscan /run/clamav \
# Clean
    && dnf clean -y all --enablerepo='*' \
    && rm -Rf /tmp/*

# Configure clamAV to run in foreground with port 3310
RUN sed -i 's/^Example$/# Example/g' /etc/clamd.d/scan.conf \
    && sed -i 's/^#Foreground .*$/Foreground true/g' /etc/clamd.d/scan.conf \
    && sed -i 's/^#TCPSocket .*$/TCPSocket 3310/g' /etc/clamd.d/scan.conf \
    && sed -i 's/^#Foreground .*$/Foreground true/g' /etc/freshclam.conf

RUN freshclam --quiet --no-dns

ADD ./server.* /etc/ssl/clamav-rest/

COPY entrypoint.sh /usr/bin/
RUN mkdir /etc/clamav/ && ln -s /etc/clamd.d/scan.conf /etc/clamav/clamd.conf

EXPOSE 9000
EXPOSE 9443

ENV MAX_SCAN_SIZE=100M
ENV MAX_FILE_SIZE=25M
ENV MAX_RECURSION=16
ENV MAX_FILES=10000
ENV MAX_EMBEDDEDPE=10M
ENV MAX_HTMLNORMALIZE=10M
ENV MAX_HTMLNOTAGS=2M
ENV MAX_SCRIPTNORMALIZE=5M
ENV MAX_ZIPTYPERCG=1M
ENV MAX_PARTITIONS=50
ENV MAX_ICONSPE=100
ENV PCRE_MATCHLIMIT=100000
ENV PCRE_RECMATCHLIMIT=2000
ENV SIGNATURE_CHECKS=24

ENTRYPOINT [ "entrypoint.sh" ]