#!/bin/bash
# 30-click-flags.sh — CLI click/hover flags (previously blocked by cobra)

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab click --wait-nav"

pt_ok nav "${FIXTURES_URL}/index.html"
pt_ok snap --interactive
# Try click with --wait-nav; element may not be a link but flag should not cause cobra error
pt click e0 --wait-nav

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab click --css"

pt_ok nav "${FIXTURES_URL}/form.html"
pt_ok click --css "button[type=submit]"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab hover (basic)"

pt_ok nav "${FIXTURES_URL}/form.html"
pt_ok snap
pt_ok hover e0

end_test
