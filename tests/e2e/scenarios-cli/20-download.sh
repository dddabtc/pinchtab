#!/bin/bash
# 20-download.sh — CLI download command

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab download (rejects private IP)"

# The download endpoint blocks private/internal IPs (SSRF protection)
pt_fail download "${FIXTURES_URL}/index.html"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab download (public URL)"

pt_ok download "https://httpbin.org/robots.txt"
assert_output_contains "data" "response contains download data"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab download (save to file)"

pt_ok download "https://httpbin.org/robots.txt" -o /tmp/e2e-download-test.txt
if [ -f /tmp/e2e-download-test.txt ]; then
  echo -e "  ${GREEN}✓${NC} file saved"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} file not saved"
  ((ASSERTIONS_FAILED++)) || true
fi
rm -f /tmp/e2e-download-test.txt

end_test
