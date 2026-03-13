#!/bin/bash
# 18-tabs-lock.sh — Tab locking operations

source "$(dirname "$0")/common.sh"

# SKIP: tabs lock/unlock not yet implemented in cobra CLI refactor
# See: feat/cli-update branch
# start_test "pinchtab tabs lock/unlock <tabId>"
# pt_ok nav "${FIXTURES_URL}/index.html"
# TAB_ID=$(echo "$PT_OUT" | jq -r '.tabId')
# pt_ok tabs lock "$TAB_ID" --owner "test-suite"
# pt_ok tabs unlock "$TAB_ID" --owner "test-suite"
# end_test
