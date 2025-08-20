#!/usr/bin/env bash
set -euo pipefail

# Linux (apt-based) Flutter + Android SDK bootstrap
# - Installs prerequisites, Java, Flutter (git clone stable), Android cmdline tools
# - Installs Android API 36/35 platforms and build-tools
# - Accepts licenses and runs flutter doctor

# Config
SDK_ROOT="${HOME}/Android/sdk"
# Load shared versions from JSON using jq
SCRIPT_DIR="$(cd -- "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
VERSIONS_JSON="${SCRIPT_DIR}/../../versions.json"
FLUTTER_DIR="${HOME}/flutter"

detect_shell_rc() {
  if [[ -n "${BASH_VERSION-}" ]]; then echo "${HOME}/.bashrc"; return; fi
  if [[ -n "${ZSH_VERSION-}" ]]; then echo "${HOME}/.zshrc"; return; fi
  echo "${HOME}/.bashrc"
}

ensure_line_in_file() {
  local line="$1" file="$2"
  grep -Fqx "$line" "$file" 2>/dev/null || echo "$line" >> "$file"
}

echo "==> Installing prerequisites (apt) ..."
sudo apt-get update -y
sudo apt-get install -y curl unzip zip git xz-utils libglu1-mesa default-jdk jq

# Parse versions.json
ANDROID_CMDLINE_TOOLS_VERSION="$(jq -r '.androidCmdlineToolsVersion' "${VERSIONS_JSON}")"
mapfile -t ANDROID_PLATFORMS < <(jq -r '.androidPlatforms[]' "${VERSIONS_JSON}")
mapfile -t ANDROID_BUILD_TOOLS < <(jq -r '.androidBuildTools[]' "${VERSIONS_JSON}")
FLUTTER_CHANNEL="$(jq -r '.flutterChannel' "${VERSIONS_JSON}")"

# Compose OS-specific cmdline tools URL using shared version token
CMDLINE_ZIP_URL="https://dl.google.com/android/repository/commandlinetools-linux-${ANDROID_CMDLINE_TOOLS_VERSION}_latest.zip"

echo "==> Installing Flutter (${FLUTTER_CHANNEL}) to ${FLUTTER_DIR} ..."
if [[ ! -d "${FLUTTER_DIR}" ]]; then
  git clone https://github.com/flutter/flutter.git -b "${FLUTTER_CHANNEL}" "${FLUTTER_DIR}"
else
  (cd "${FLUTTER_DIR}" && git fetch && git checkout "${FLUTTER_CHANNEL}" && git pull)
fi
export PATH="${FLUTTER_DIR}/bin:${PATH}"

echo "==> Preparing Android SDK directories at ${SDK_ROOT} ..."
mkdir -p "${SDK_ROOT}/cmdline-tools" "${SDK_ROOT}/platform-tools"

echo "==> Downloading Android command-line tools ..."
TMP_DIR="$(mktemp -d)"
curl -fsSL "${CMDLINE_ZIP_URL}" -o "${TMP_DIR}/cmdline-tools.zip"
unzip -q "${TMP_DIR}/cmdline-tools.zip" -d "${TMP_DIR}"
rm -rf "${SDK_ROOT}/cmdline-tools/latest" || true
mkdir -p "${SDK_ROOT}/cmdline-tools/latest"
cp -R "${TMP_DIR}/cmdline-tools/"* "${SDK_ROOT}/cmdline-tools/latest/"
rm -rf "${TMP_DIR}"

export ANDROID_SDK_ROOT="${SDK_ROOT}"
export ANDROID_HOME="${SDK_ROOT}"
export PATH="${SDK_ROOT}/cmdline-tools/latest/bin:${SDK_ROOT}/platform-tools:${PATH}"

echo "==> Accepting Android SDK licenses ..."
yes | sdkmanager --licenses >/dev/null || true

echo "==> Installing Android platform-tools, platforms and build-tools ..."
sdkmanager --install "platform-tools" >/dev/null || true
for p in "${ANDROID_PLATFORMS[@]}"; do
  sdkmanager --install "platforms;${p}" >/dev/null || true
done
for b in "${ANDROID_BUILD_TOOLS[@]}"; do
  sdkmanager --install "build-tools;${b}" >/dev/null || true
done

RC_FILE="$(detect_shell_rc)"
echo "==> Updating shell profile: ${RC_FILE}"
ensure_line_in_file "export ANDROID_SDK_ROOT=\"${SDK_ROOT}\"" "${RC_FILE}"
ensure_line_in_file "export ANDROID_HOME=\"${SDK_ROOT}\"" "${RC_FILE}"
ensure_line_in_file "export PATH=\"${SDK_ROOT}/cmdline-tools/latest/bin:\$PATH\"" "${RC_FILE}"
ensure_line_in_file "export PATH=\"${SDK_ROOT}/platform-tools:\$PATH\"" "${RC_FILE}"
ensure_line_in_file "export PATH=\"${FLUTTER_DIR}/bin:\$PATH\"" "${RC_FILE}"

echo "==> Running flutter doctor ..."
flutter doctor || true

echo "✅ Linux setup complete. Open a new shell or 'source' your RC file to refresh PATH."
