#!/bin/bash
# 23-redirects.sh — Redirect following via CLI

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "redirects: follow single redirect"

pt_ok nav "https://httpbin.org/redirect/1"
pt_ok snap
assert_json_field_contains ".url" "httpbin.org/get" "landed on /get after redirect"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "redirects: follow multiple redirects"

pt_ok nav "https://httpbin.org/redirect/3"
pt_ok snap
assert_json_field_contains ".url" "httpbin.org/get" "multiple redirects followed to /get"

end_test
