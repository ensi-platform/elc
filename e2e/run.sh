#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
E2E_DIR="${ROOT_DIR}/e2e"
COMPOSE_FILE="${E2E_DIR}/docker-compose.yml"
BINARY_PATH="${ROOT_DIR}/build/elc"
IMAGES_FILE="${E2E_DIR}/images.txt"
CACHE_DIR="${E2E_DIR}/.image-cache"

cd "${ROOT_DIR}"

cleanup() {
  docker compose -f "${COMPOSE_FILE}" down -v --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT

image_cache_name() {
  local image="$1"
  echo "${image}" | tr '/:' '__'
}

# Pull missing images on the host and refresh tar cache when image IDs change.
ensure_image_cache() {
  mkdir -p "${CACHE_DIR}"

  if [[ ! -f "${IMAGES_FILE}" ]]; then
    echo "ERROR: missing ${IMAGES_FILE}" >&2
    exit 1
  fi

  while IFS= read -r image || [[ -n "${image}" ]]; do
    image="${image%%#*}"
    image="$(echo "${image}" | xargs)"
    [[ -z "${image}" ]] && continue

    local safe tar stamp host_id cached_id
    safe="$(image_cache_name "${image}")"
    tar="${CACHE_DIR}/${safe}.tar"
    stamp="${CACHE_DIR}/${safe}.id"
    host_id=""

    if docker image inspect "${image}" >/dev/null 2>&1; then
      host_id="$(docker image inspect -f '{{.Id}}' "${image}")"
    fi

    if [[ -f "${tar}" && -f "${stamp}" ]]; then
      cached_id="$(cat "${stamp}")"
      if [[ -n "${host_id}" && "${host_id}" != "${cached_id}" ]]; then
        echo "  refresh: ${image}"
        docker save -o "${tar}" "${image}"
        printf '%s\n' "${host_id}" >"${stamp}"
      else
        echo "  hit: ${image}"
      fi
      continue
    fi

    if [[ -z "${host_id}" ]]; then
      echo "  pull: ${image}"
      docker pull "${image}"
      host_id="$(docker image inspect -f '{{.Id}}' "${image}")"
    fi

    echo "  save: ${image}"
    docker save -o "${tar}" "${image}"
    printf '%s\n' "${host_id}" >"${stamp}"
  done <"${IMAGES_FILE}"
}

load_images_into_dind() {
  docker compose -f "${COMPOSE_FILE}" exec -T e2e sh -c '
    set -e
    if ! ls /image-cache/*.tar >/dev/null 2>&1; then
      echo "ERROR: no image tarballs in /image-cache" >&2
      exit 1
    fi
    for tar in /image-cache/*.tar; do
      echo "  load: $(basename "$tar")"
      docker load -i "$tar" >/dev/null
    done
  '
}

echo "==> build linux elc binary"
mkdir -p "${ROOT_DIR}/build"
# Remove a stale path that Docker may have created as a directory when the binary was missing.
if [[ -d "${BINARY_PATH}" ]]; then
  rm -rf "${BINARY_PATH}"
fi

VERSION="$(./version.sh)"
GOOS=linux GOARCH="$(go env GOARCH)" \
  go build -o "${BINARY_PATH}" \
  -ldflags="-X 'github.com/ensi-platform/elc/core.Version=${VERSION}'" \
  main.go
chmod +x "${BINARY_PATH}"

echo "==> prepare image cache"
ensure_image_cache

echo "==> start DinD"
docker compose -f "${COMPOSE_FILE}" up -d --build --wait

echo "==> load images into DinD"
load_images_into_dind

echo "==> run smoke"
docker compose -f "${COMPOSE_FILE}" exec -T e2e /e2e/tests/smoke.sh

echo "==> e2e OK"
