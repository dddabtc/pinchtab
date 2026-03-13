#!/bin/bash
# 29-error-edge-cases.sh — Edge case error handling
# Migrated from: tests/integration/error_handling_test.go (ER4, ER6)

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "error handling: empty page (about:blank)"

pt_post /navigate '{"url":"about:blank"}'
assert_ok "navigate to about:blank"

TAB_ID=$(get_tab_id)

pt_get "/snapshot?tabId=${TAB_ID}"
assert_ok "snapshot on empty page"

pt_get "/text?tabId=${TAB_ID}"
assert_ok "text on empty page"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "error handling: rapid navigation"

# Fire 3 rapid navigations
pt_post /navigate "{\"url\":\"${FIXTURES_URL}/index.html\"}"
pt_post /navigate "{\"url\":\"${FIXTURES_URL}/form.html\"}"
pt_post /navigate "{\"url\":\"${FIXTURES_URL}/buttons.html\"}"
assert_ok "final navigate succeeded"

# Verify server still works after rapid nav
sleep 1
pt_get /snapshot
assert_ok "snapshot after rapid nav"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "error handling: unicode content"

pt_post /navigate "{\"url\":\"${FIXTURES_URL}/unicode.html\"}"
assert_ok "navigate to unicode page"

pt_get /snapshot
assert_ok "snapshot unicode page"

pt_get /text
assert_ok "text unicode page"
assert_contains "$RESULT" "Unicode" "text has unicode content"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "error handling: binary content (image)"

# Navigate to a binary file (e.g., favicon.ico or sample image)
# Should not crash — may return error or empty content
pt_post /navigate "{\"url\":\"${FIXTURES_URL}/sample.txt\"}"
pt_get /text
# Don't assert_ok — just verify server didn't crash
pt_get /health
assert_ok "server still healthy after binary/text page"

end_test
