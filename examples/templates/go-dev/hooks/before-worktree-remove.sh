#!/usr/bin/env bash
set -euo pipefail

# runs before worktree folder is deleted; write marker outside the worktree
marker_dir="${HOME_PATH:-${WORKSPACE_PATH}/home}"
mkdir -p "${marker_dir}"
marker="${marker_dir}/.elc-worktree-remove-${APP_NAME}"

echo "worktree_remove: ${APP_NAME} branch=${GIT_BRANCH:-?} -> ${SVC_PATH}"
printf '%s\n' "${GIT_BRANCH:-}" >"${marker}"
