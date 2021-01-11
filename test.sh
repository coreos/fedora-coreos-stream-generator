#!/bin/bash
set -euo pipefail
outf=$(mktemp --suffix=fcos-stream)
expected=fixtures/stream.json
./fedora-coreos-stream-generator -pretty-print -releases ./fixtures/releases.json -output-file "${outf}".tmp
# The last-modified changes based on time
jq "del(.metadata)" < "${outf}".tmp > "${outf}" && rm -f "${outf}".tmp
if ! diff -u "${expected}" "${outf}"; then
    echo "error: Failed to match expected ${expected}" 1>&2
    exit 1
fi
rm -f "${outf}"
echo "ok"
