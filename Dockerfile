FROM alpine
MAINTAINER Niilo Ursin <niilo.ursin+nospam_github@gmail.com>

RUN apk --no-cache add clamav clamav-libunrar \
    && mkdir /run/clamav \
    && chown clamav:clamav /run/clamav

RUN sed -i 's/^#Foreground .*$/Foreground true/g' /etc/clamav/clamd.conf \
    && sed -i 's/^#TCPSocket .*$/TCPSocket 3310/g' /etc/clamav/clamd.conf \
    && sed -i 's/^#Foreground .*$/Foreground true/g' /etc/clamav/freshclam.conf

RUN freshclam --quiet

COPY entrypoint.sh /usr/bin/
COPY clamav-rest /usr/bin/

EXPOSE 9000

ENTRYPOINT [ "entrypoint.sh" ]