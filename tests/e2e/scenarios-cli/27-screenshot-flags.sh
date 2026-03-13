#!/bin/bash
# 27-screenshot-flags.sh — CLI screenshot flags

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab screenshot -o custom.jpg"

pt_ok nav "${FIXTURES_URL}/index.html"
pt_ok screenshot -o /tmp/e2e-custom-screenshot.jpg

if [ -f /tmp/e2e-custom-screenshot.jpg ]; then
  echo -e "  ${GREEN}✓${NC} file created"
  ((ASSERTIONS_PASSED++)) || true
  rm -f /tmp/e2e-custom-screenshot.jpg
else
  echo -e "  ${RED}✗${NC} file not created"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab screenshot -q 10"

pt_ok screenshot -q 10 -o /tmp/e2e-lowq.jpg
rm -f /tmp/e2e-lowq.jpg

end_test
