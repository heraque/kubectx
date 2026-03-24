#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_BIN_DIR="${ROOT_DIR}/bin"
TARGET_BIN_DIR=""
INSTALL_OS=""
INSTALL_ARCH=""
PATH_LINE='export PATH="$HOME/.local/bin:$PATH"'
ALIAS_LINE='alias ctx="kubectx"'
RC_FILE=""
PATH_FILES=("${HOME}/.zprofile" "${HOME}/.zshrc" "${HOME}/.zshenv" "${HOME}/.bashrc" "${HOME}/.profile")

usage() {
  cat <<'EOF'
Uso:
  ./install.sh [--os darwin|linux] [--arch amd64|arm64] [--target-dir DIR]

Padroes:
  macOS  -> ~/.local/bin
  Linux  -> /usr/local/bin
EOF
}

normalize_os() {
  case "${1}" in
    Darwin|darwin) printf 'darwin\n' ;;
    Linux|linux) printf 'linux\n' ;;
    *)
      echo "sistema operacional nao suportado: ${1}" >&2
      exit 1
      ;;
  esac
}

normalize_arch() {
  case "${1}" in
    x86_64|amd64) printf 'amd64\n' ;;
    arm64|aarch64) printf 'arm64\n' ;;
    *)
      echo "arquitetura nao suportada: ${1}" >&2
      exit 1
      ;;
  esac
}

detect_rc_file() {
  local shell_name
  shell_name="$(basename "${SHELL:-}")"

  if [[ "${INSTALL_OS}" == "darwin" ]]; then
    printf '%s\n' "${HOME}/.zprofile"
    return 0
  fi

  case "${shell_name}" in
    zsh)
      printf '%s\n' "${HOME}/.zshrc"
      ;;
    bash)
      printf '%s\n' "${HOME}/.bashrc"
      ;;
    *)
      printf '%s\n' "${HOME}/.bashrc"
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
      echo "argumento desconhecido: ${1}" >&2
      usage
      exit 1
      ;;
  esac
done

INSTALL_OS="${INSTALL_OS:-$(normalize_os "$(uname -s)")}"
INSTALL_ARCH="${INSTALL_ARCH:-$(normalize_arch "$(uname -m)")}"
TARGET_BIN_DIR="${TARGET_BIN_DIR:-$([[ "${INSTALL_OS}" == "linux" ]] && printf '/usr/local/bin' || printf '%s/.local/bin' "${HOME}")}"
SOURCE_KUBECTX="${SOURCE_BIN_DIR}/${INSTALL_OS}-${INSTALL_ARCH}/kubectx"
RC_FILE="$(detect_rc_file)"

mkdir -p "${TARGET_BIN_DIR}"

if [[ ! -x "${SOURCE_KUBECTX}" ]]; then
  echo "binario kubectx ausente para ${INSTALL_OS}/${INSTALL_ARCH}: ${SOURCE_KUBECTX}" >&2
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
    } >> "${RC_FILE}"
  fi
fi

FOUND_CTX_ALIAS=0
for path_file in "${PATH_FILES[@]}"; do
  [[ -f "${path_file}" ]] || continue
  if grep -F 'alias ctx=' "${path_file}" >/dev/null 2>&1; then
    FOUND_CTX_ALIAS=1
    break
  fi
done

if [[ "${FOUND_CTX_ALIAS}" -eq 0 ]]; then
  {
    printf '\n'
    printf '%s\n' '# Added by local-setup/install.sh'
    printf '%s\n' "${ALIAS_LINE}"
  } >> "${RC_FILE}"
fi

printf 'Plataforma detectada: %s/%s\n' "${INSTALL_OS}" "${INSTALL_ARCH}"
printf 'Arquivo de shell rc: %s\n' "${RC_FILE}"
printf 'kubectx instalado em %s\n' "${TARGET_BIN_DIR}/kubectx"
printf 'wrapper kubectl instalado em %s\n' "${TARGET_BIN_DIR}/kubectl"
if [[ "${TARGET_BIN_DIR}" == "${HOME}/.local/bin" ]]; then
  printf 'Abra um novo shell ou recarregue %s\n' "${RC_FILE}"
fi
