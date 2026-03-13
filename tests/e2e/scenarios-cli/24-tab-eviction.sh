#!/bin/bash
# 24-tab-eviction.sh — LRU tab eviction via CLI
#
# The default test instance has maxTabs=10.
# We use nav to create tabs within the same instance,
# then verify eviction when exceeding the limit.

source "$(dirname "$0")/common.sh"

MAX_TABS=10

# ─────────────────────────────────────────────────────────────────
start_test "tab eviction: open tabs up to limit"

TAB_IDS=()

for i in $(seq 1 $MAX_TABS); do
  pt_ok nav "${FIXTURES_URL}/index.html?t=$i"
  TAB_IDS+=($(echo "$PT_OUT" | jq -r '.tabId'))
done

# Verify all our tabs exist
pt_ok tabs
TAB_COUNT=$(echo "$PT_OUT" | jq '.tabs | length')
if [ "$TAB_COUNT" -ge "$MAX_TABS" ]; then
  echo -e "  ${GREEN}✓${NC} $TAB_COUNT tabs open (>= $MAX_TABS)"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} expected >= $MAX_TABS tabs, got $TAB_COUNT"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "tab eviction: new tab evicts oldest"

FIRST_TAB="${TAB_IDS[0]}"
TABS_BEFORE=$(echo "$PT_OUT" | jq '.tabs | length')

# Open one more — should trigger eviction of the oldest
sleep 1
pt_ok nav "${FIXTURES_URL}/index.html?t=overflow"

# Verify the oldest tab was evicted
pt_ok tabs
TABS_JSON="$PT_OUT"

if echo "$TABS_JSON" | grep -q "$FIRST_TAB"; then
  echo -e "  ${RED}✗${NC} oldest tab should have been evicted"
  ((ASSERTIONS_FAILED++)) || true
else
  echo -e "  ${GREEN}✓${NC} oldest tab evicted (LRU)"
  ((ASSERTIONS_PASSED++)) || true
fi

# Verify tab count didn't grow (eviction kept it in check)
TABS_AFTER=$(echo "$TABS_JSON" | jq '.tabs | length')
if [ "$TABS_AFTER" -le "$TABS_BEFORE" ]; then
  echo -e "  ${GREEN}✓${NC} tab count stable after eviction ($TABS_AFTER <= $TABS_BEFORE)"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} tab count grew from $TABS_BEFORE to $TABS_AFTER"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test
