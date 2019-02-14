#!/bin/sh

echo "$(date) - $CLAM_VIRUSEVENT_VIRUSNAME > $CLAM_VIRUSEVENT_FILENAME" >> /var/log/clamav-infected.log
if [ -e "$CLAM_VIRUSEVENT_FILENAME" ] && [ ! -d  "$CLAM_VIRUSEVENT_FILENAME" ]; then
    echo "Virus $CLAM_VIRUSEVENT_FILENAME exist and is not a directory"
    # rm $CLAM_VIRUSEVENT_FILENAME
fi


# curl {API}