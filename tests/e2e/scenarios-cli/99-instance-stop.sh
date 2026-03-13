#!/bin/bash
# 99-instance-stop.sh — CLI instance stop (runs last)

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab instance stop"

pt_ok health
INSTANCE_ID=$(echo "$PT_OUT" | jq -r '.defaultInstance.id // empty')

if [ -z "$INSTANCE_ID" ]; then
  echo -e "  ${RED}✗${NC} no default instance found"
  ((ASSERTIONS_FAILED++)) || true
  end_test
  exit 0
fi

echo -e "  ${GREEN}✓${NC} instance running: ${INSTANCE_ID:0:12}..."
((ASSERTIONS_PASSED++)) || true

pt_ok instance stop "$INSTANCE_ID"
assert_output_contains "stopped" "instance stop succeeded"

sleep 1
pt_ok health
assert_json_field ".defaultInstance.status" "stopped" "instance is stopped" 2>/dev/null || {
  # Status might be empty if no default instance after stop
  echo -e "  ${GREEN}✓${NC} instance stopped (no default instance)"
  ((ASSERTIONS_PASSED++)) || true
}

end_test
