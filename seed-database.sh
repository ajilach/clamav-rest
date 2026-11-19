#!/bin/bash
# Makes sure a ClamAV database directory contains a usable database, using the snapshot bundled with
# the image where needed. Without a database clamd refuses to start.
#
# Usage: seed-database.sh <bundled database dir> <database dir>
#
# The bundled copy of a database (main, daily, bytecode) is used if the database directory
# - has no .cvd or .cld file of it; other files like lost+found, hidden files or a freshclam.dat left over
#   from a failed update are ignored,
# - has one without a valid header, or a .cvd that fails verification, e.g. after an interrupted copy, or
# - has an older version than the bundled one, e.g. a persistent volume that freshclam could not update
#   because the CDN blocks or rate-limits the node.
# A .cld is never replaced because of the verification: freshclam creates it by patching a .cvd and it has
# no signature of its own. Only a newer bundled version replaces it.

src=$1
dst=$2

if [ -z "$src" ] || [ -z "$dst" ]; then
    echo "Usage: $0 <bundled database dir> <database dir>" >&2
    exit 2
fi

# Prints the version from the header of a .cvd/.cld file ("ClamAV-VDB:<date>:<version>:..."),
# or nothing if the header can't be read.
db_version() {
    local header version
    header=$(head -c 512 "$1" 2>/dev/null | tr -d '\0')
    case "$header" in
        ClamAV-VDB:*) ;;
        *) return ;;
    esac
    version=$(echo "$header" | cut -d: -f3)
    if [[ "$version" =~ ^[0-9]+$ ]]; then
        echo "$version"
    fi
}

if ! mkdir -p "$dst" || [ ! -w "$dst" ]; then
    echo "ERROR: the database directory $dst does not exist or is not writable for user $(id -un), clamd will not be able to start without a database." >&2
    exit 1
fi

status=0
found_bundled=false
for bundled in "$src"/*.cvd; do
    [ -e "$bundled" ] || continue
    found_bundled=true
    name=${bundled##*/}
    name=${name%.cvd}
    bundled_version=$(db_version "$bundled")

    # freshclam keeps either a .cld or a .cvd of a database, the .cld being the patched, newer one
    existing=""
    for f in "$dst/$name.cld" "$dst/$name.cvd"; do
        if [ -s "$f" ]; then
            existing=$f
            break
        fi
    done

    reason=""
    if [ -z "$existing" ]; then
        reason="there is no $name.cvd or $name.cld in $dst"
    else
        existing_version=$(db_version "$existing")
        if [ -z "$existing_version" ]; then
            reason="$existing has no valid header"
        elif [[ "$existing" == *.cvd ]] && ! sigtool --info "$existing" 2>/dev/null | grep -q "^Verification OK"; then
            reason="$existing failed verification"
        elif [ -n "$bundled_version" ] && [ "$bundled_version" -gt "$existing_version" ]; then
            reason="$existing has the older version $existing_version"
        fi
    fi
    [ -n "$reason" ] || continue

    echo "Using the bundled $name database (version ${bundled_version:-unknown}), because $reason."
    # copy under a temporary name first, so an interrupted copy never leaves a broken database behind
    tmp="$dst/.$name.cvd.seed"
    if cp "$bundled" "$tmp" && mv -f "$tmp" "$dst/$name.cvd"; then
        # an older .cld of the same database would otherwise be kept next to the new .cvd
        if [ -e "$dst/$name.cld" ]; then
            rm -f "$dst/$name.cld"
        fi
    else
        rm -f "$tmp"
        echo "ERROR: unable to copy $bundled to $dst" >&2
        status=1
    fi
done

if [ "$found_bundled" = false ]; then
    echo "WARNING: there are no bundled databases in $src" >&2
fi

exit $status
