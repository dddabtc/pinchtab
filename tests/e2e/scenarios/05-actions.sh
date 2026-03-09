#!/bin/bash
# 05-actions.sh — Browser actions (click, type, press)

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab click <button>"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/buttons.html\"}"
sleep 1

pt_get /snapshot
click_button "Increment"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab type <field> <text>"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/form.html\"}"
sleep 1

pt_get /snapshot
type_into "Username" "testuser123"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab press <key>"

press_key "Escape"

end_test
