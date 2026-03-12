#!/bin/bash
# 18-tabs-lock.sh — Tab locking operations

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab tabs lock/unlock <tabId>"

pt_ok nav "${FIXTURES_URL}/index.html"
TAB_ID=$(echo "$PT_OUT" | jq -r '.tabId')

# Lock the tab with owner
pt_ok tabs lock "$TAB_ID" --owner "test-suite"
assert_output_contains "locked" "tab locked successfully"

# Unlock the tab (needs same owner)
pt_ok tabs unlock "$TAB_ID" --owner "test-suite"
assert_output_contains "unlocked" "tab unlocked successfully"

end_test
