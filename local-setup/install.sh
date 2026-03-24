#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_BIN_DIR="${ROOT_DIR}/bin"
TARGET_BIN_DIR=""
INSTALL_OS=""
INSTALL_ARCH=""
ZPROFILE="${HOME}/.zprofile"
PATH_LINE='export PATH="$HOME/.local/bin:$PATH"'
PATH_FILES=("${HOME}/.zprofile" "${HOME}/.zshrc" "${HOME}/.zshenv" "${HOME}/.profile")

usage() {
  cat <<'EOF'
Usage:
  ./install.sh [--os darwin|linux] [--arch amd64|arm64] [--target-dir DIR]

Defaults:
  macOS  -> ~/.local/bin
  Linux  -> /usr/local/bin
EOF
}

normalize_os() {
  case "${1}" in
    Darwin|darwin) printf 'darwin\n' ;;
    Linux|linux) printf 'linux\n' ;;
    *)
      echo "unsupported OS: ${1}" >&2
      exit 1
      ;;
  esac
}

normalize_arch() {
  case "${1}" in
    x86_64|amd64) printf 'amd64\n' ;;
    arm64|aarch64) printf 'arm64\n' ;;
    *)
      echo "unsupported architecture: ${1}" >&2
      exit 1
      ;;
  esac
}

while [[ $# -gt 0 ]]; do
  case "${1}" in
    --os)
      INSTALL_OS="$(normalize_os "${2}")"
      shift 2
      ;;
    --arch)
      INSTALL_ARCH="$(normalize_arch "${2}")"
      shift 2
      ;;
    --target-dir)
      TARGET_BIN_DIR="${2}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: ${1}" >&2
      usage
      exit 1
      ;;
  esac
done

INSTALL_OS="${INSTALL_OS:-$(normalize_os "$(uname -s)")}"
INSTALL_ARCH="${INSTALL_ARCH:-$(normalize_arch "$(uname -m)")}"
TARGET_BIN_DIR="${TARGET_BIN_DIR:-$([[ "${INSTALL_OS}" == "linux" ]] && printf '/usr/local/bin' || printf '%s/.local/bin' "${HOME}")}"
SOURCE_KUBECTX="${SOURCE_BIN_DIR}/${INSTALL_OS}-${INSTALL_ARCH}/kubectx"

mkdir -p "${TARGET_BIN_DIR}"

if [[ ! -x "${SOURCE_KUBECTX}" ]]; then
  echo "missing kubectx binary for ${INSTALL_OS}/${INSTALL_ARCH}: ${SOURCE_KUBECTX}" >&2
  exit 1
fi

install -m 0755 "${SOURCE_KUBECTX}" "${TARGET_BIN_DIR}/kubectx"
install -m 0755 "${SOURCE_BIN_DIR}/kubectl" "${TARGET_BIN_DIR}/kubectl"

if [[ "${TARGET_BIN_DIR}" == "${HOME}/.local/bin" ]]; then
  FOUND_PATH_ENTRY=0
  for path_file in "${PATH_FILES[@]}"; do
    [[ -f "${path_file}" ]] || continue
    if grep -F '.local/bin' "${path_file}" >/dev/null 2>&1; then
      FOUND_PATH_ENTRY=1
      break
    fi
  done

  if [[ "${FOUND_PATH_ENTRY}" -eq 0 ]]; then
    {
      printf '\n'
      printf '%s\n' '# Added by local-setup/install.sh'
      printf '%s\n' "${PATH_LINE}"
    } >> "${ZPROFILE}"
  fi
fi

printf 'Detected platform: %s/%s\n' "${INSTALL_OS}" "${INSTALL_ARCH}"
printf 'Installed kubectx to %s\n' "${TARGET_BIN_DIR}/kubectx"
printf 'Installed kubectl wrapper to %s\n' "${TARGET_BIN_DIR}/kubectl"
if [[ "${TARGET_BIN_DIR}" == "${HOME}/.local/bin" ]]; then
  printf 'Open a new shell or run: exec zsh -l\n'
fi
