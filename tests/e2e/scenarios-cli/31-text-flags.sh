#!/bin/bash
# 31-text-flags.sh — CLI text flags (previously blocked by cobra)

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab text --raw"

pt_ok nav "${FIXTURES_URL}/index.html"
pt_ok text --raw

end_test
