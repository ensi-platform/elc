#!/usr/bin/env bash
set -euo pipefail

# runs after `elc clone` for components using this hook
marker_dir="${HOME_PATH:-${WORKSPACE_PATH}/home}"
mkdir -p "${marker_dir}"

echo "after_clone: ${APP_NAME} -> ${SVC_PATH}"
touch "${marker_dir}/.elc-after-clone-${APP_NAME}"
