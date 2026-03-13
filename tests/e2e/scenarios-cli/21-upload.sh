#!/bin/bash
# 21-upload.sh — CLI upload command

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab upload (basic)"

pt_ok nav "${FIXTURES_URL}/upload.html"

echo "test content" > /tmp/e2e-upload-test.txt
pt_ok upload /tmp/e2e-upload-test.txt --selector "#single-file"
assert_output_contains "ok" "upload succeeded"
rm -f /tmp/e2e-upload-test.txt

end_test
