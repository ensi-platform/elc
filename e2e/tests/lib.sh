#!/usr/bin/env bash

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local message="${3:-expected to find '${needle}'}"

  if [[ "${haystack}" != *"${needle}"* ]]; then
    echo "ASSERT FAILED: ${message}" >&2
    echo "--- output ---" >&2
    echo "${haystack}" >&2
    return 1
  fi
}

assert_file_exists() {
  local path="$1"
  local message="${2:-expected file '${path}' to exist}"

  if [[ ! -e "${path}" ]]; then
    echo "ASSERT FAILED: ${message}" >&2
    return 1
  fi
}

assert_container_running() {
  local name_fragment="$1"
  local ps_output

  ps_output="$(docker ps --format '{{.Names}}')"
  assert_contains "${ps_output}" "${name_fragment}" "container matching '${name_fragment}' is not running"
}

setup_workspace_copy() {
  local dest="$1"

  rm -rf "${dest}"
  mkdir -p "${dest}"
  cp -a /examples/. "${dest}/"
  rm -rf "${dest}/home"
  mkdir -p "${dest}/home"
}

ensure_network() {
  local network_name="${1:-example}"

  if ! docker network inspect "${network_name}" >/dev/null 2>&1; then
    docker network create "${network_name}" >/dev/null
  fi
}

init_git_repo() {
  local repo_path="$1"

  git -C "${repo_path}" init -b main >/dev/null
  git -C "${repo_path}" config user.email "e2e@example.com"
  git -C "${repo_path}" config user.name "e2e"
  git -C "${repo_path}" add -A
  git -C "${repo_path}" commit -m "init" >/dev/null
}

create_bare_remote() {
  local src_path="$1"
  local bare_path="$2"

  rm -rf "${bare_path}"
  git clone --bare "${src_path}" "${bare_path}" >/dev/null
  git -C "${src_path}" remote remove origin >/dev/null 2>&1 || true
  git -C "${src_path}" remote add origin "${bare_path}"
  git -C "${src_path}" push -u origin main >/dev/null
}

# Inject repository for go-api and append a cloneable service that uses go-dev hooks.
prepare_workspace_for_git_commands() {
  local ws_path="$1"
  local go_bare="$2"
  local clone_bare="$3"
  local yaml="${ws_path}/workspace.yaml"
  local tmp
  local clone_src

  init_git_repo "${ws_path}/apps/go-api"
  create_bare_remote "${ws_path}/apps/go-api" "${go_bare}"

  clone_src="$(mktemp -d)"
  echo "cloneable" >"${clone_src}/README"
  init_git_repo "${clone_src}"
  create_bare_remote "${clone_src}" "${clone_bare}"
  rm -rf "${clone_src}"

  tmp="$(mktemp)"
  awk -v go_repo="${go_bare}" -v clone_repo="${clone_bare}" '
    /^  go-api:/ { in_go = 1 }
    in_go && /^    path:/ {
      print
      print "    repository: " go_repo
      in_go = 0
      next
    }
    /^modules:/ {
      print "  cloneable:"
      print "    path: ${APPS_ROOT}/cloneable"
      print "    repository: " clone_repo
      print "    compose_file: null"
      print "    hooks:"
      print "      after_clone: ${WORKSPACE_PATH}/templates/go-dev/hooks/after-clone.sh"
      print ""
    }
    { print }
  ' "${yaml}" >"${tmp}"
  mv "${tmp}" "${yaml}"
}
