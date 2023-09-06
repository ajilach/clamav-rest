#!/bin/bash

# Replace values with environment variables in clamd.conf
sed -i 's/^#MaxScanSize .*$/MaxScanSize '"$MAX_SCAN_SIZE"'/g' /etc/clamav/clamd.conf
sed -i 's/^#StreamMaxLength .*$/StreamMaxLength '"$MAX_FILE_SIZE"'/g' /etc/clamav/clamd.conf
sed -i 's/^#MaxFileSize .*$/MaxFileSize '"$MAX_FILE_SIZE"'/g' /etc/clamav/clamd.conf
sed -i 's/^#MaxRecursion .*$/MaxRecursion '"$MAX_RECURSION"'/g' /etc/clamav/clamd.conf
sed -i 's/^#MaxFiles .*$/MaxFiles '"$MAX_FILES"'/g' /etc/clamav/clamd.conf
sed -i 's/^#MaxEmbeddedPE .*$/MaxEmbeddedPE '"$MAX_EMBEDDEDPE"'/g' /etc/clamav/clamd.conf
sed -i 's/^#MaxHTMLNormalize .*$/MaxHTMLNormalize '"$MAX_HTMLNORMALIZE"'/g' /etc/clamav/clamd.conf
sed -i 's/^#MaxHTMLNoTags.*$/MaxHTMLNoTags '"$MAX_HTMLNOTAGS"'/g' /etc/clamav/clamd.conf
sed -i 's/^#MaxScriptNormalize .*$/MaxScriptNormalize '"$MAX_SCRIPTNORMALIZE"'/g' /etc/clamav/clamd.conf
sed -i 's/^#MaxZipTypeRcg .*$/MaxZipTypeRcg '"$MAX_ZIPTYPERCG"'/g' /etc/clamav/clamd.conf
sed -i 's/^#MaxPartitions .*$/MaxPartitions '"$MAX_PARTITIONS"'/g' /etc/clamav/clamd.conf
sed -i 's/^#MaxIconsPE .*$/MaxIconsPE '"$MAX_ICONSPE"'/g' /etc/clamav/clamd.conf
sed -i 's/^#PCREMatchLimit.*$/PCREMatchLimit '"$PCRE_MATCHLIMIT"'/g' /etc/clamav/clamd.conf
sed -i 's/^#PCRERecMatchLimit .*$/PCRERecMatchLimit '"$PCRE_RECMATCHLIMIT"'/g' /etc/clamav/clamd.conf

if [ -n "$PROXY_SERVER" ]; then
    sed -i 's~^#HTTPProxyServer .*~HTTPProxyServer '"$PROXY_SERVER"'~g' /etc/clamav/freshclam.conf

    # It's not required, but if they also provided a port, then configure it
    if [ -n "$PROXY_PORT" ]; then
        sed -i 's/^#HTTPProxyPort .*$/HTTPProxyPort '"$PROXY_PORT"'/g' /etc/clamav/freshclam.conf
    fi

    # It's not required, but if they also provided a username, then configure both the username and password
    if [ -n "$PROXY_USERNAME" ]; then
        sed -i 's/^#HTTPProxyUsername .*$/HTTPProxyUsername '"$PROXY_USERNAME"'/g' /etc/clamav/freshclam.conf
        sed -i 's~^#HTTPProxyPassword .*~HTTPProxyPassword '"$PROXY_PASSWORD"'~g' /etc/clamav/freshclam.conf
    fi
fi

(
    freshclam --daemon --checks=$SIGNATURE_CHECKS &
    clamd &
    /usr/bin/clamav-rest &
) 2>&1 | tee -a /var/log/clamav/clamav.log


pids=`jobs -p`

exitcode=0

terminate() {
    for pid in $pids; do
        if ! kill -0 $pid 2>/dev/null; then
            wait $pid
            exitcode=$?
        fi
    done
    kill $pids 2>/dev/null
}

trap terminate CHLD
wait

exit $exitcode
