#!/bin/bash
runTests=false
if [ "$1" = "test" ]; then
    runTests=true
fi
cp /etc/clamav/* /clamav/etc/

# Replace values in freshclam.conf
sed -i 's/^#\?NotifyClamd .*$/NotifyClamd \/clamav\/etc\/clamd.conf/g' /clamav/etc/freshclam.conf
sed -i 's/^#DatabaseDirectory .*$/DatabaseDirectory \/clamav\/data/g' /clamav/etc/freshclam.conf
sed -i 's/^#\?NotifyClamd .*$/NotifyClamd \/clamav\/etc\/clamd.conf/g' /clamav/etc/freshclam.conf
sed -i 's/^#TemporaryDirectory .*$/TemporaryDirectory \/clamav\/tmp/g' /clamav/etc/clamd.conf
sed -i 's/^#DatabaseDirectory .*$/DatabaseDirectory \/clamav\/data/g' /clamav/etc/clamd.conf

# Replace values with environment variables in freshclam.conf
sed -i 's/^#\?Checks .*$/Checks '"$SIGNATURE_CHECKS"'/g' /clamav/etc/freshclam.conf

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

#define function to terminate the container
terminate () {
    pids=`jobs -p`
    for pid in $pids; do
        if [ "$$" != "$pid" ]; then
            if ! kill -0 $pid 2>/dev/null; then
                wait -n $pid
                exitcode=$?
            fi
        fi
    done
    kill $pids 2>/dev/null
}

if [ -z "$(ls -A /clamav/data)" ]; then
  cp /var/lib/clamav/* /clamav/data/
fi
exitcode=0
output=$(
    freshclam --config-file=/clamav/etc/freshclam.conf --daemon &
    clamd --config-file=/clamav/etc/clamd.conf &
    /usr/bin/clamav-rest &
    
    if [ "$runTests" == true ]; then
        echo "Will run test and then exit"
        /opt/clamav-rest/run-tests 
        # the exit code from `run-tests` is the numberOfFailedSteps
        res=$?
        # terminate the other processes of the container.
        terminate
        # the value of $exitcode set in the terminate function is not used because we want 
        # the result of the tests to determine the exit of the container when running tests.
        # If no tests have failed, the exit code should be 0
        exitcode=$res
        echo "number of failed test and exit code: $exitcode"
        exit $exitcode
    else 
        # Force reload the virus database through the clamd socket after 120s.
        # Starting freshclam and clamd async ends up that a newer database version is loaded with
        # freshclam, but the clamd still keep the old version existing before the update because 
        # the socket from clamd is not yet ready to inform, what is indicated in the log
        # during the startup of the container (WARNING: Clamd was NOT notified: Can't connect to clamd through /run/clamav/clamd.sock: No such file or directory).
        # So only if a newer database version is available clamd will be notified next time, and this can take hours/days.
        # Remarks: The socket port is configured in the .Dockerfile itself.
        sleep 120s
        echo RELOAD | nc 127.0.0.01 3310 &
    fi
)
# Capture exit code from the sub shell to use it when exiting the container.    
exitcode=$? 

echo "$output" | tee -a /var/log/clamav/clamav.log
# if we have run tests, don't terminate CHLD, just exit.
if [ "$runTests" == true ]; then
    exit $exitcode
fi
trap terminate CHLD
echo "waiting to finish"
wait

exit $exitcode
