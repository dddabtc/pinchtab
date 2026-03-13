#!/bin/bash
# 25-find.sh — CLI find command

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find (basic)"

pt_ok nav "${FIXTURES_URL}/form.html"
pt_ok find "username"
assert_output_contains "ref" "has ref in output"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find --ref-only"

pt_ok find "username" --ref-only
# Output should be just a ref like "e5"
assert_output_contains "e" "outputs ref"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find --explain"

pt_ok find "submit" --explain

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab find (no match)"

pt find "xyznonexistent99999"
# May succeed with low score or fail - just verify it doesn't crash

end_test
