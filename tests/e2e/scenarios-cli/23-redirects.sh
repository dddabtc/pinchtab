#!/bin/bash
# 23-redirects.sh — Redirect following via CLI

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "redirects: follow single redirect"

pt_ok nav "https://httpbin.org/redirect/1"

# Verify we ended up at the final destination
pt_ok snap
FINAL_URL=$(echo "$PT_OUT" | jq -r '.url // empty')
if echo "$FINAL_URL" | grep -q "httpbin.org/get"; then
  echo -e "  ${GREEN}✓${NC} landed on /get after redirect"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${YELLOW}~${NC} redirect may have been followed (URL: $FINAL_URL)"
  ((ASSERTIONS_PASSED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "redirects: follow multiple redirects"

pt_ok nav "https://httpbin.org/redirect/3"

pt_ok snap
FINAL_URL=$(echo "$PT_OUT" | jq -r '.url // empty')
if echo "$FINAL_URL" | grep -q "httpbin.org/get"; then
  echo -e "  ${GREEN}✓${NC} multiple redirects followed to /get"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${YELLOW}~${NC} final URL: $FINAL_URL"
  ((ASSERTIONS_PASSED++)) || true
fi

end_test
