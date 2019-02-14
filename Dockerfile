FROM golang:alpine

# Update
RUN apk update upgrade;

# Set timezone to Singapore
RUN apk add tzdata
RUN mv /etc/localtime /etc/localtime.utc && \
    ln -s /usr/share/zoneinfo/Asia/Singapore /etc/localtime


# Install ClamAV
RUN apk --no-cache add clamav clamav-libunrar \
    && mkdir /run/clamav \
    && chown clamav:clamav /run/clamav

# Configure clamAV to run in foreground with port 3310
RUN sed -i 's/^#Foreground .*$/Foreground true/g' /etc/clamav/clamd.conf \
    && sed -i 's/^#TCPSocket .*$/TCPSocket 3310/g' /etc/clamav/clamd.conf \
    && sed -i 's/^#Foreground .*$/Foreground true/g' /etc/clamav/freshclam.conf


# Configure scan alerts
RUN touch /var/log/clamav-infected.log && chmod 0777 /var/log/clamav-infected.log
COPY ./alert.sh /opt/clamav-utils/
RUN chmod -Rf 0755 /opt/clamav-utils/alert.sh
RUN sed -i 's:^#VirusEvent .*$:VirusEvent /opt/clamav-utils/alert.sh:g' /etc/clamd.d/scan.conf


# Configure ClamAV user, ScanOnAccess requires root
#RUN sed -i 's/^User .*$/User root/g' /etc/clamd.d/scan.conf


# ScanOnAccess configurations
RUN mkdir /scan-target
RUN sed -i 's/^#ScanOnAccess .*$/ScanOnAccess yes/g' /etc/clamav/clamd.conf \
    && sed -i 's:#OnAccessMountPath /home/user:&\r\nOnAccessMountPath /scan-target:g' /etc/clamav/clamd.conf


RUN freshclam -v --no-dns
# --quiet

# Build go package
ADD . /go/src/clamav-rest/
RUN cd /go/src/clamav-rest && go build -v


COPY entrypoint.sh /usr/bin/
RUN mv /go/src/clamav-rest/clamav-rest /usr/bin/ && rm -Rf /go/src/clamav-rest


EXPOSE 9000

ENTRYPOINT [ "entrypoint.sh" ]