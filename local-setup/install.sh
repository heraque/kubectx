#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_BIN_DIR="${ROOT_DIR}/bin"
TARGET_BIN_DIR="${HOME}/.local/bin"
ZPROFILE="${HOME}/.zprofile"
PATH_LINE='export PATH="$HOME/.local/bin:$PATH"'
PATH_FILES=("${HOME}/.zprofile" "${HOME}/.zshrc" "${HOME}/.zshenv" "${HOME}/.profile")

mkdir -p "${TARGET_BIN_DIR}"

install -m 0755 "${SOURCE_BIN_DIR}/kubectx" "${TARGET_BIN_DIR}/kubectx"
install -m 0755 "${SOURCE_BIN_DIR}/kubectl" "${TARGET_BIN_DIR}/kubectl"

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

printf 'Installed kubectx to %s\n' "${TARGET_BIN_DIR}/kubectx"
printf 'Installed kubectl wrapper to %s\n' "${TARGET_BIN_DIR}/kubectl"
printf 'Open a new shell or run: exec zsh -l\n'
