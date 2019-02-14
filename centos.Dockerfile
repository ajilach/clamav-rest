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



# Set timezone to Singapore
RUN mv /etc/localtime /etc/localtime.utc && \
    ln -s /usr/share/zoneinfo/Asia/Singapore /etc/localtime



# Configure clamAV to run in foreground with port 3310
RUN sed -i 's/^Example$/# Example/g' /etc/clamd.d/scan.conf \
    && sed -i 's/^#Foreground .*$/Foreground true/g' /etc/clamd.d/scan.conf \
    && sed -i 's/^#TCPSocket .*$/TCPSocket 3310/g' /etc/clamd.d/scan.conf \
    && sed -i 's/^#Foreground .*$/Foreground true/g' /etc/freshclam.conf


# Configure scan alerts
COPY ./alert.sh /opt/clamav-utils/
RUN touch /var/log/clamav-infected.log && chmod 0777 /var/log/clamav-infected.log
RUN chmod -Rf 0755 /opt/clamav-utils/alert.sh
RUN sed -i 's:^#VirusEvent .*$:VirusEvent /opt/clamav-utils/alert.sh:g' /etc/clamd.d/scan.conf


# Configure ClamAV user, ScanOnAccess requires root
#RUN sed -i 's/^User .*$/User root/g' /etc/clamd.d/scan.conf


# ScanOnAccess configurations
RUN mkdir /scan-target
RUN sed -i 's/^#ScanOnAccess .*$/ScanOnAccess yes/g' /etc/clamd.d/scan.conf \
    && sed -i 's:#OnAccessMountPath /home/user:&\r\nOnAccessMountPath /scan-target:g' /etc/clamd.d/scan.conf


# Build go package
ADD . /go/src/clamav-rest/
RUN cd /go/src/clamav-rest/ && go build -v


COPY entrypoint.sh /usr/bin/
RUN mv /go/src/clamav-rest/clamav-rest /usr/bin/ && rm -Rf /go/src/clamav-rest


EXPOSE 9000



RUN freshclam -v --no-dns
# --quiet


ENTRYPOINT [ "entrypoint.sh" ]