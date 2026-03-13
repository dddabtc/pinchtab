#!/bin/bash
# 29-nav-flags.sh — CLI nav flags (previously blocked by cobra)

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab nav --new-tab"

pt_ok nav "${FIXTURES_URL}/index.html"
pt_ok nav "${FIXTURES_URL}/form.html" --new-tab

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab nav --block-images"

pt_ok nav "${FIXTURES_URL}/index.html" --block-images

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab nav --block-ads"

pt_ok nav "${FIXTURES_URL}/index.html" --block-ads

end_test
