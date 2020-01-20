FROM centos

RUN yum -y update && yum clean all

# Install golang
RUN mkdir -p /go && chmod -R 777 /go && \
    yum install -y centos-release-scl epel-release && \
    yum -y install golang nano && yum clean all

ENV GOPATH=/go \
    PATH="$GOPATH/bin:/usr/local/go/bin:$PATH"

# Install ClamAV
RUN yum install -y clamav-server clamav-data clamav-update clamav-filesystem clamav clamav-scanner-systemd clamav-devel clamav-lib clamav-server-systemd \
    && mkdir /run/clamav \
    && chown clamscan:clamscan /run/clamav

# Clean
RUN yum clean -y all --enablerepo='*' && \
    rm -Rf /tmp/*

# Set timezone to Europe/Zurich
RUN ln -s /usr/share/zoneinfo/Europe/Zurich /etc/localtime

# Configure clamAV to run in foreground with port 3310
RUN sed -i 's/^Example$/# Example/g' /etc/clamd.d/scan.conf \
    && sed -i 's/^#Foreground .*$/Foreground true/g' /etc/clamd.d/scan.conf \
    && sed -i 's/^#TCPSocket .*$/TCPSocket 3310/g' /etc/clamd.d/scan.conf \
    && sed -i 's/^#Foreground .*$/Foreground true/g' /etc/freshclam.conf

# Build go package
ADD . /go/src/clamav-rest/
RUN cd /go/src/clamav-rest/ && go build -v

COPY entrypoint.sh /usr/bin/
RUN mv /go/src/clamav-rest/clamav-rest /usr/bin/ && rm -Rf /go/src/clamav-rest

EXPOSE 9000

RUN freshclam --quiet

ENTRYPOINT [ "entrypoint.sh" ]