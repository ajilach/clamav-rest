#!/bin/bash
# Tests for seed-database.sh. Runs during the build of Dockerfile.test, where /var/lib/clamav holds the
# databases bundled with the image. Must not run as root, otherwise the permission test is meaningless.
#
# Usage: seed-database_test.sh <path to seed-database.sh> <bundled database dir>

seed=$1
bundled=$2
failures=0
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

fail() {
    echo "FAIL: $*"
    failures=$((failures + 1))
}

pass() {
    echo "PASS: $*"
}

# writes a file that only consists of a database header with the given version
fake_db() {
    printf 'ClamAV-VDB:01 Jan 2020 00-00 +0000:%s:1:90:00000000000000000000000000000000:x:test:1577836800' "$2" > "$1"
}

same_file() {
    cmp -s "$1" "$2"
}

daily_version=$(head -c 512 "$bundled/daily.cvd" | cut -d: -f3)

# 1) other files but no database
dir=$work/1
mkdir -p "$dir/lost+found" && touch "$dir/.hidden" && echo x > "$dir/freshclam.dat"
if "$seed" "$bundled" "$dir" >/dev/null &&
    same_file "$bundled/main.cvd" "$dir/main.cvd" &&
    same_file "$bundled/daily.cvd" "$dir/daily.cvd" &&
    same_file "$bundled/bytecode.cvd" "$dir/bytecode.cvd" &&
    [ -d "$dir/lost+found" ] && [ -e "$dir/.hidden" ] && [ "$(cat "$dir/freshclam.dat")" = x ]; then
    pass "seeds all databases if the directory only contains other files"
else
    fail "seeds all databases if the directory only contains other files"
fi

# 2) only some databases
dir=$work/2
mkdir -p "$dir" && cp "$bundled/main.cvd" "$dir/" && touch -d "2020-01-01" "$dir/main.cvd"
if "$seed" "$bundled" "$dir" >/dev/null &&
    [ -e "$dir/daily.cvd" ] && [ -e "$dir/bytecode.cvd" ] &&
    [ "$(date -r "$dir/main.cvd" +%Y)" = 2020 ]; then
    pass "only seeds missing databases"
else
    fail "only seeds missing databases"
fi

# 3) newer .cld is kept, it is not signature-verified
dir=$work/3
mkdir -p "$dir" && cp "$bundled/main.cvd" "$bundled/bytecode.cvd" "$dir/"
fake_db "$dir/daily.cld" $((daily_version + 1))
"$seed" "$bundled" "$dir" >/dev/null
if [ "$(head -c 512 "$dir/daily.cld" | cut -d: -f3)" = $((daily_version + 1)) ] && [ ! -e "$dir/daily.cvd" ]; then
    pass "keeps a newer .cld"
else
    fail "keeps a newer .cld"
fi

# 4) older .cld is replaced by the bundled .cvd
dir=$work/4
mkdir -p "$dir" && cp "$bundled/main.cvd" "$bundled/bytecode.cvd" "$dir/"
fake_db "$dir/daily.cld" $((daily_version - 1))
output=$("$seed" "$bundled" "$dir")
if same_file "$bundled/daily.cvd" "$dir/daily.cvd" && [ ! -e "$dir/daily.cld" ] && [[ "$output" == *"older version"* ]]; then
    pass "replaces an older database"
else
    fail "replaces an older database: $output"
fi

# 5) truncated and invalid .cvd files are replaced
dir=$work/5
mkdir -p "$dir" && cp "$bundled/bytecode.cvd" "$dir/"
head -c 1000000 "$bundled/daily.cvd" > "$dir/daily.cvd"
echo "not a database" > "$dir/main.cvd"
output=$("$seed" "$bundled" "$dir")
if same_file "$bundled/daily.cvd" "$dir/daily.cvd" && same_file "$bundled/main.cvd" "$dir/main.cvd" &&
    [[ "$output" == *"failed verification"* ]] && [[ "$output" == *"no valid header"* ]]; then
    pass "replaces truncated and invalid databases"
else
    fail "replaces truncated and invalid databases: $output"
fi

# 6) up-to-date databases are left alone and nothing is logged
dir=$work/6
mkdir -p "$dir" && cp "$bundled"/*.cvd "$dir/"
output=$("$seed" "$bundled" "$dir")
if [ -z "$output" ]; then
    pass "does nothing if all databases are up to date"
else
    fail "does nothing if all databases are up to date: $output"
fi

# 7) missing directory is created
dir=$work/7/data
if "$seed" "$bundled" "$dir" >/dev/null && [ -e "$dir/daily.cvd" ]; then
    pass "creates a missing database directory"
else
    fail "creates a missing database directory"
fi

# 8) directory that is not writable
dir=$work/8
mkdir -p "$dir" && chmod 555 "$dir"
output=$("$seed" "$bundled" "$dir" 2>&1)
status=$?
chmod 755 "$dir"
if [ $status -ne 0 ] && [[ "$output" == *"not writable"* ]]; then
    pass "fails with a clear error if the directory is not writable"
else
    fail "fails with a clear error if the directory is not writable (exit $status): $output"
fi

if [ $failures -ne 0 ]; then
    echo "$failures seed-database test(s) failed"
    exit 1
fi
echo "All seed-database tests passed"
