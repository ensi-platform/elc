#!/usr/bin/env bash
set -euo pipefail

# runs after `elc worktree` / `elc worktree add`; args after `--` are forwarded here
# markers go to HOME_PATH so the worktree stays clean for `worktree remove`
marker_dir="${HOME_PATH:-${WORKSPACE_PATH}/home}"
mkdir -p "${marker_dir}"
marker_prefix="${marker_dir}/.elc-worktree-create-${APP_NAME}"

echo "worktree_create: ${APP_NAME} branch=${GIT_BRANCH:-?} -> ${SVC_PATH}"
if [[ "$#" -gt 0 ]]; then
  echo "worktree_create args: $*"
  printf '%s\n' "$@" >"${marker_prefix}-args"
fi
touch "${marker_prefix}"
