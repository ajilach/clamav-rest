#!/bin/bash

mkdir -p /clamav/etc
mkdir -p /clamav/data
mkdir -p /clamav/tmp
cp /etc/clamav/* /clamav/etc/

sed -i 's/^#DatabaseDirectory .*$/DatabaseDirectory \/clamav\/data/g' /clamav/etc/freshclam.conf
sed -i 's/^#TemporaryDirectory .*$/TemporaryDirectory \/clamav\/tmp/g' /clamav/etc/clamd.conf
sed -i 's/^#DatabaseDirectory .*$/DatabaseDirectory \/clamav\/data/g' /clamav/etc/clamd.conf

# Replace values with environment variables in clamd.conf
sed -i 's/^#MaxScanSize .*$/MaxScanSize '"$MAX_SCAN_SIZE"'/g' /clamav/etc/clamd.conf
sed -i 's/^#StreamMaxLength .*$/StreamMaxLength '"$MAX_FILE_SIZE"'/g' /clamav/etc/clamd.conf
sed -i 's/^#MaxFileSize .*$/MaxFileSize '"$MAX_FILE_SIZE"'/g' /clamav/etc/clamd.conf
sed -i 's/^#MaxRecursion .*$/MaxRecursion '"$MAX_RECURSION"'/g' /clamav/etc/clamd.conf
sed -i 's/^#MaxFiles .*$/MaxFiles '"$MAX_FILES"'/g' /clamav/etc/clamd.conf
sed -i 's/^#MaxEmbeddedPE .*$/MaxEmbeddedPE '"$MAX_EMBEDDEDPE"'/g' /clamav/etc/clamd.conf
sed -i 's/^#MaxHTMLNormalize .*$/MaxHTMLNormalize '"$MAX_HTMLNORMALIZE"'/g' /clamav/etc/clamd.conf
sed -i 's/^#MaxHTMLNoTags.*$/MaxHTMLNoTags '"$MAX_HTMLNOTAGS"'/g' /clamav/etc/clamd.conf
sed -i 's/^#MaxScriptNormalize .*$/MaxScriptNormalize '"$MAX_SCRIPTNORMALIZE"'/g' /clamav/etc/clamd.conf
sed -i 's/^#MaxZipTypeRcg .*$/MaxZipTypeRcg '"$MAX_ZIPTYPERCG"'/g' /clamav/etc/clamd.conf
sed -i 's/^#MaxPartitions .*$/MaxPartitions '"$MAX_PARTITIONS"'/g' /clamav/etc/clamd.conf
sed -i 's/^#MaxIconsPE .*$/MaxIconsPE '"$MAX_ICONSPE"'/g' /clamav/etc/clamd.conf
sed -i 's/^#PCREMatchLimit.*$/PCREMatchLimit '"$PCRE_MATCHLIMIT"'/g' /clamav/etc/clamd.conf
sed -i 's/^#PCRERecMatchLimit .*$/PCRERecMatchLimit '"$PCRE_RECMATCHLIMIT"'/g' /clamav/etc/clamd.conf

if [ -z "$(ls -A /clamav/data)" ]; then
  cp /var/lib/clamav/* /clamav/data/
fi

(
    freshclam --config-file=/clamav/etc/freshclam.conf --daemon --checks=$SIGNATURE_CHECKS &
    clamd --config-file=/clamav/etc/clamd.conf &
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
