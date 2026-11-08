#!/bin/bash

FRESHCLAM_CONF=/clamav/etc/freshclam.conf
LOG_FILE=/var/log/clamav/clamav.log

# Messages of this script also go to the log file, like the output of the services started below.
log() {
    echo "$*" | tee -a "$LOG_FILE" 2>/dev/null
}
log_error() {
    echo "$*" | tee -a "$LOG_FILE" >&2 2>/dev/null
}

# conf_has <option> [<value>...] checks whether freshclam.conf sets the option to one of the values, or to
# any value if none are given. Option names are case-sensitive, like in freshclam.
conf_has() {
    awk -v opt="$1" -v vals="${*:2}" '
        BEGIN { n = split(vals, v, " ") }
        $1 == opt && NF >= 2 {
            val = $2
            gsub(/^"|"$/, "", val)
            if (n == 0) found = 1
            for (i = 1; i <= n; i++) if (val == v[i]) found = 1
        }
        END { exit !found }' "$FRESHCLAM_CONF"
}

# CUSTOM_FRESHCLAM_CONF decides whether an existing /clamav/etc/freshclam.conf is provided by the user and
# must be left untouched:
#   auto (default)  custom if the file is mounted into the container (docker bind mount, Kubernetes
#                   ConfigMap) or is not writable. A copy written by this script on an earlier start
#                   (persistent /clamav volume) is neither, so it is regenerated.
#   true            always keep an existing file, e.g. one edited directly on a persistent volume
#   false           always regenerate the file from the defaults and the environment variables
custom_freshclam_conf=false
case "${CUSTOM_FRESHCLAM_CONF:-auto}" in
    true)
        if [ -e "$FRESHCLAM_CONF" ]; then
            custom_freshclam_conf=true
        else
            log "CUSTOM_FRESHCLAM_CONF is true, but there is no $FRESHCLAM_CONF yet. Using the default configuration."
        fi
        ;;
    false)
        ;;
    auto)
        if [ -e "$FRESHCLAM_CONF" ]; then
            if awk -v p="$FRESHCLAM_CONF" '$5 == p { found = 1 } END { exit !found }' /proc/self/mountinfo; then
                custom_freshclam_conf=true
            elif [ ! -w "$FRESHCLAM_CONF" ]; then
                custom_freshclam_conf=true
                log "$FRESHCLAM_CONF is not writable and therefore treated as custom. Set CUSTOM_FRESHCLAM_CONF=false to regenerate it."
            fi
        fi
        ;;
    *)
        log_error "ERROR: CUSTOM_FRESHCLAM_CONF must be auto, true or false, got '$CUSTOM_FRESHCLAM_CONF'."
        exit 1
        ;;
esac

for f in /etc/clamav/*; do
    if [ "$custom_freshclam_conf" = true ] && [ "${f##*/}" = "freshclam.conf" ]; then
        continue
    fi
    # -f replaces files that are not writable, e.g. a copy owned by another user from an earlier start
    cp -f "$f" /clamav/etc/ || log_error "ERROR: unable to copy $f to /clamav/etc/"
done

# Replace values in clamd.conf
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

if [ "$custom_freshclam_conf" = true ]; then
    log "Using custom $FRESHCLAM_CONF, SIGNATURE_CHECKS and PROXY_* environment variables are ignored."
    if [ ! -r "$FRESHCLAM_CONF" ]; then
        log_error "ERROR: custom $FRESHCLAM_CONF is not readable for user $(id -un), freshclam will not be able to start."
    else
        conf_has Foreground yes true 1 ||
            log_error "WARNING: custom $FRESHCLAM_CONF does not set 'Foreground yes', which is required for the container to work correctly."
        conf_has DatabaseDirectory /clamav/data /clamav/data/ ||
            log_error "WARNING: custom $FRESHCLAM_CONF does not set 'DatabaseDirectory /clamav/data', which is required for the container to work correctly."
        conf_has NotifyClamd /clamav/etc/clamd.conf ||
            log_error "WARNING: custom $FRESHCLAM_CONF does not set 'NotifyClamd /clamav/etc/clamd.conf', which is required for the container to work correctly."
        conf_has DatabaseMirror || conf_has PrivateMirror ||
            log_error "WARNING: custom $FRESHCLAM_CONF contains neither DatabaseMirror nor PrivateMirror, freshclam will not be able to update the database."
    fi
else
    # Replace values in freshclam.conf
    sed -i 's/^#\?NotifyClamd .*$/NotifyClamd \/clamav\/etc\/clamd.conf/g' "$FRESHCLAM_CONF"
    sed -i 's/^#DatabaseDirectory .*$/DatabaseDirectory \/clamav\/data/g' "$FRESHCLAM_CONF"

    # Replace values with environment variables in freshclam.conf
    sed -i 's/^#\?Checks .*$/Checks '"$SIGNATURE_CHECKS"'/g' "$FRESHCLAM_CONF"

    # Use proxy for freshclam
    if [ -n "$PROXY_SERVER" ]; then
        sed -i 's~^#HTTPProxyServer .*~HTTPProxyServer '"$PROXY_SERVER"'~g' "$FRESHCLAM_CONF"

        # It's not required, but if they also provided a port, then configure it
        if [ -n "$PROXY_PORT" ]; then
            sed -i 's/^#HTTPProxyPort .*$/HTTPProxyPort '"$PROXY_PORT"'/g' "$FRESHCLAM_CONF"
        fi

        # It's not required, but if they also provided a username, then configure both the username and password
        if [ -n "$PROXY_USERNAME" ]; then
            if [ -z "$PROXY_PASSWORD" ]; then
                log_error "ERROR: PROXY_PASSWORD must be set when PROXY_USERNAME is provided."
                exit 1
            fi
            sed -i 's/^#HTTPProxyUsername .*$/HTTPProxyUsername '"$PROXY_USERNAME"'/g' "$FRESHCLAM_CONF"
            sed -i 's~^#HTTPProxyPassword .*~HTTPProxyPassword '"$PROXY_PASSWORD"'~g' "$FRESHCLAM_CONF"
            # Freshclam requires 0700 permissions if a password is set in the config.
            chmod 0700 "$FRESHCLAM_CONF"
        fi
    fi
fi

if [ -z "$(ls -A /clamav/data)" ]; then
  cp /var/lib/clamav/* /clamav/data/
fi

(
    freshclam --config-file=/clamav/etc/freshclam.conf --daemon &
    clamd --config-file=/clamav/etc/clamd.conf &
    /usr/bin/clamav-rest &
    # Force reload the virus database through the clamd socket after 120s.
    # Starting freshclam and clamd async ends up that a newer database version is loaded with
    # freshclam, but the clamd still keep the old version existing before the update because 
    # the socket from clamd is not yet ready to inform, what is indicated in the log
    # during the startup of the container (WARNING: Clamd was NOT notified: Can't connect to clamd through /run/clamav/clamd.sock: No such file or directory).
    # So only if a newer database version is available clamd will be notified next time, and this can take hours/days.
    # Remarks: The socket port is configured in the .Dockerfile itself.
    sleep 120s
    echo RELOAD | nc 127.0.0.01 3310 &
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
