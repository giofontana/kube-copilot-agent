#!/usr/bin/env bash
# setup-project-board.sh — Creates the "KubeCopilot Roadmap" GitHub project board,
# adds all issues and feature PRs, and sets their Kanban status.
#
# Prerequisites:
#   gh auth refresh -s project,read:project
#
# Usage:
#   ./hack/setup-project-board.sh

set -euo pipefail

OWNER="giofontana"
REPO="kube-copilot-agent"
REPO_URL="https://github.com/${OWNER}/${REPO}"

echo "==> Creating project board..."
PROJECT_JSON=$(gh project create --owner "$OWNER" --title "KubeCopilot Roadmap" --format json)
PROJECT_NUMBER=$(echo "$PROJECT_JSON" | jq -r '.number')
echo "    Project #${PROJECT_NUMBER} created"

echo "==> Adding issues to the project..."
for num in 3 4 5 6 11 13 15 16 17 18 19 20 21 22 23; do
  echo "    Adding issue #${num}"
  gh project item-add "$PROJECT_NUMBER" --owner "$OWNER" \
    --url "${REPO_URL}/issues/${num}" 2>/dev/null || echo "    (already added or error)"
done

echo "==> Adding feature PRs to the project..."
for num in 7 8 9 10 12 14; do
  echo "    Adding PR #${num}"
  gh project item-add "$PROJECT_NUMBER" --owner "$OWNER" \
    --url "${REPO_URL}/pull/${num}" 2>/dev/null || echo "    (already added or error)"
done

echo "==> Fetching field metadata..."
FIELDS_JSON=$(gh project field-list "$PROJECT_NUMBER" --owner "$OWNER" --format json)
STATUS_FIELD_ID=$(echo "$FIELDS_JSON" | jq -r '.fields[] | select(.name=="Status") | .id')

if [ -z "$STATUS_FIELD_ID" ] || [ "$STATUS_FIELD_ID" = "null" ]; then
  echo "    WARNING: No Status field found. Items added but statuses not set."
  echo "    Project URL: https://github.com/users/${OWNER}/projects/${PROJECT_NUMBER}"
  exit 0
fi

TODO_OPT=$(echo "$FIELDS_JSON" | jq -r '.fields[] | select(.name=="Status") | .options[] | select(.name=="Todo") | .id')
IN_PROGRESS_OPT=$(echo "$FIELDS_JSON" | jq -r '.fields[] | select(.name=="Status") | .options[] | select(.name=="In Progress") | .id')
DONE_OPT=$(echo "$FIELDS_JSON" | jq -r '.fields[] | select(.name=="Status") | .options[] | select(.name=="Done") | .id')

echo "    Status field: ${STATUS_FIELD_ID}"
echo "    Todo option:  ${TODO_OPT}"
echo "    In Progress:  ${IN_PROGRESS_OPT}"
echo "    Done option:  ${DONE_OPT}"

# Get the GraphQL project ID (needed for item-edit)
PROJECT_ID=$(gh project view "$PROJECT_NUMBER" --owner "$OWNER" --format json | jq -r '.id')

echo "==> Fetching item list..."
ITEMS_JSON=$(gh project item-list "$PROJECT_NUMBER" --owner "$OWNER" --format json --limit 100)

set_status() {
  local content_number="$1"
  local option_id="$2"
  local label="$3"

  local item_id
  item_id=$(echo "$ITEMS_JSON" | jq -r --arg num "$content_number" \
    '.items[] | select(.content.number == ($num | tonumber)) | .id')

  if [ -n "$item_id" ] && [ "$item_id" != "null" ]; then
    echo "    Setting #${content_number} → ${label}"
    gh project item-edit \
      --project-id "$PROJECT_ID" \
      --id "$item_id" \
      --field-id "$STATUS_FIELD_ID" \
      --single-select-option-id "$option_id" 2>/dev/null || echo "    (failed to set status)"
  else
    echo "    WARNING: Item #${content_number} not found in project"
  fi
}

if [ -n "$IN_PROGRESS_OPT" ] && [ "$IN_PROGRESS_OPT" != "null" ]; then
  echo "==> Setting In Progress statuses..."
  for num in 3 4 5 6 11 13 7 8 9 10 12 14; do
    set_status "$num" "$IN_PROGRESS_OPT" "In Progress"
  done
fi

if [ -n "$TODO_OPT" ] && [ "$TODO_OPT" != "null" ]; then
  echo "==> Setting Todo statuses..."
  for num in 15 16 17 18 19 20 21 22 23; do
    set_status "$num" "$TODO_OPT" "Todo"
  done
fi

echo ""
echo "=== Done ==="
echo "Project #${PROJECT_NUMBER}: https://github.com/users/${OWNER}/projects/${PROJECT_NUMBER}"
echo ""
echo "To update ROADMAP.md with the project link, run:"
echo "  sed -i 's|PROJECT_NUMBER|${PROJECT_NUMBER}|g' ROADMAP.md"
