#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "${SCRIPT_DIR}/lib.sh"

WS_NAME="e2e-examples"
WS_PATH="/tmp/ws-examples"
NETWORK_NAME="example"
GO_BARE="/tmp/e2e-go-api.git"
CLONE_BARE="/tmp/e2e-cloneable.git"
WT_BRANCH="e2e-smoke"

ELC=(elc --workspace="${WS_NAME}")

cleanup() {
  set +e
  "${ELC[@]}" destroy fpm >/dev/null 2>&1
  "${ELC[@]}" --component=go-api worktree remove "${WT_BRANCH}" --force --no-hook >/dev/null 2>&1
  "${ELC[@]}" destroy go-api >/dev/null 2>&1
  elc workspace remove "${WS_NAME}" >/dev/null 2>&1
  rm -rf "${WS_PATH}" "${GO_BARE}" "${CLONE_BARE}"
}
trap cleanup EXIT

echo "==> elc --version / --help"
version_out="$(elc --version)"
assert_contains "${version_out}" "." "elc --version should print a version"
help_out="$(elc --help)"
assert_contains "${help_out}" "Available Commands" "elc --help should list commands"

echo "==> prepare workspace copy"
setup_workspace_copy "${WS_PATH}"
ensure_network "${NETWORK_NAME}"
prepare_workspace_for_git_commands "${WS_PATH}" "${GO_BARE}" "${CLONE_BARE}"

echo "==> workspace add / list / select / show / set-root"
elc workspace add "${WS_NAME}" "${WS_PATH}"

list_out="$(elc workspace list)"
assert_contains "${list_out}" "${WS_NAME}" "workspace list should include ${WS_NAME}"

elc workspace select "${WS_NAME}"
show_out="$(elc workspace show)"
assert_contains "${show_out}" "${WS_NAME}" "workspace show should print ${WS_NAME}"

elc workspace set-root "${WS_NAME}" "${WS_PATH}"

echo "==> list"
list_svc_out="$("${ELC[@]}" list)"
assert_contains "${list_svc_out}" "fpm" "list should include fpm"
assert_contains "${list_svc_out}" "go-api" "list should include go-api"
assert_contains "${list_svc_out}" "cloneable" "list should include cloneable"

echo "==> vars"
vars_out="$("${ELC[@]}" vars fpm)"
assert_contains "${vars_out}" "APP_NAME=" "vars should print computed variables"
assert_contains "${vars_out}" "NETWORK=" "vars should include NETWORK"

echo "==> wrap"
wrap_out="$("${ELC[@]}" --component=fpm wrap printenv APP_NAME)"
assert_contains "${wrap_out}" "fpm" "wrap should expose component env"

echo "==> launch"
launch_out="$("${ELC[@]}" --component=go-api launch pwd)"
assert_contains "${launch_out}" "${WS_PATH}/apps/go-api" "launch should run in component directory"

echo "==> clone (+ after_clone hook)"
clone_out="$("${ELC[@]}" clone cloneable)"
assert_contains "${clone_out}" "after_clone:" "clone should run after_clone hook"
assert_file_exists "${WS_PATH}/apps/cloneable/README" "clone should create cloneable component"
assert_file_exists "${WS_PATH}/home/.elc-after-clone-cloneable" "after_clone hook should create marker"

echo "==> run-hook after_clone"
rm -f "${WS_PATH}/home/.elc-after-clone-cloneable"
run_hook_out="$("${ELC[@]}" --component=cloneable run-hook after_clone)"
assert_contains "${run_hook_out}" "after_clone:" "run-hook should execute after_clone hook"
assert_file_exists "${WS_PATH}/home/.elc-after-clone-cloneable" "run-hook after_clone should recreate marker"

echo "==> worktree add / list / remove (+ hooks)"
wt_add_out="$("${ELC[@]}" --component=go-api worktree add "${WT_BRANCH}" --source=HEAD -- --env=staging)"
assert_contains "${wt_add_out}" "worktree_create:" "worktree add should run worktree_create hook"
assert_file_exists "${WS_PATH}/worktrees/go-api/${WT_BRANCH}" "worktree add should create worktree path"
assert_file_exists "${WS_PATH}/home/.elc-worktree-create-go-api-${WT_BRANCH}" "worktree_create hook should create marker"
assert_contains "$(cat "${WS_PATH}/home/.elc-worktree-create-go-api-${WT_BRANCH}-args")" "--env=staging" "worktree_create should receive args after --"

wt_list_out="$("${ELC[@]}" worktree list)"
assert_contains "${wt_list_out}" "go-api" "worktree list should include component"
assert_contains "${wt_list_out}" "${WT_BRANCH}" "worktree list should include branch"

wt_rm_out="$("${ELC[@]}" --component=go-api worktree remove "${WT_BRANCH}")"
assert_contains "${wt_rm_out}" "worktree_remove:" "worktree remove should run worktree_remove hook"
assert_file_exists "${WS_PATH}/home/.elc-worktree-remove-go-api-${WT_BRANCH}" "worktree_remove hook should create marker"
assert_contains "$(cat "${WS_PATH}/home/.elc-worktree-remove-go-api-${WT_BRANCH}")" "${WT_BRANCH}" "worktree_remove marker should contain branch"
if [[ -e "${WS_PATH}/worktrees/go-api/${WT_BRANCH}" ]]; then
  echo "ASSERT FAILED: worktree path still exists after remove" >&2
  exit 1
fi

# After worktree ops: set-hooks installs wrappers for every git hook name, which
# would otherwise break subsequent git worktree commands in this repo.
echo "==> set-hooks"
mkdir -p "${WS_PATH}/apps/go-api/e2e-hooks/pre-commit"
printf '#!/bin/sh\necho ok\n' >"${WS_PATH}/apps/go-api/e2e-hooks/pre-commit/check.sh"
chmod +x "${WS_PATH}/apps/go-api/e2e-hooks/pre-commit/check.sh"
"${ELC[@]}" --component=go-api set-hooks e2e-hooks
assert_file_exists "${WS_PATH}/apps/go-api/.git/hooks/pre-commit" "set-hooks should generate pre-commit wrapper"

echo "==> start fpm (+ proxy dependency)"
"${ELC[@]}" start fpm
assert_container_running "elc-examples-proxy"
assert_container_running "elc-examples-fpm"

echo "==> compose"
compose_out="$("${ELC[@]}" --component=fpm compose ps)"
assert_contains "${compose_out}" "app" "compose ps should mention app service"

echo "==> exec"
exec_out="$("${ELC[@]}" --component=fpm exec --uid=0 --no-tty pwd)"
assert_contains "${exec_out}" "/" "exec should run command in container"

echo "==> run"
run_out="$("${ELC[@]}" --component=fpm run --uid=0 --no-tty echo smoke-run)"
assert_contains "${run_out}" "smoke-run" "run should execute command in a new container"

echo "==> restart"
"${ELC[@]}" restart fpm
assert_container_running "elc-examples-fpm"

echo "==> stop"
"${ELC[@]}" stop fpm

echo "==> destroy"
"${ELC[@]}" destroy fpm

echo "==> smoke OK"
