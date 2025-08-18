#!/usr/bin/env bash
set -euo pipefail

# macOS Flutter + Android SDK bootstrap
# - Installs Temurin (OpenJDK), Flutter
# - Installs Android command-line tools (no full Android Studio)
# - Installs Android API 36/35 platforms and build-tools
# - Accepts licenses and runs flutter doctor

# Config
SDK_ROOT="${HOME}/Library/Android/sdk"
CMDLINE_ZIP_URL="https://dl.google.com/android/repository/commandlinetools-mac-11076708_latest.zip"
BUILD_TOOLS=("36.0.0" "35.0.0")
PLATFORMS=("android-36" "android-35")

detect_shell_rc() {
  if [[ -n "${ZSH_VERSION-}" ]]; then echo "${HOME}/.zshrc"; return; fi
  if [[ -n "${BASH_VERSION-}" ]]; then echo "${HOME}/.bash_profile"; return; fi
  # Default to zsh on modern macOS
  echo "${HOME}/.zshrc"
}

ensure_line_in_file() {
  local line="$1" file="$2"
  grep -Fqx "$line" "$file" 2>/dev/null || echo "$line" >> "$file"
}

echo "==> Ensuring Homebrew..."
if ! command -v brew >/dev/null 2>&1; then
  echo "Homebrew not found. Installing (requires sudo) ..."
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi

echo "==> Installing Temurin (OpenJDK) and Flutter..."
brew install --cask temurin >/dev/null || true
brew install --cask flutter >/dev/null || true

# Ensure flutter is on PATH for this session
if ! command -v flutter >/dev/null 2>&1; then
  if [[ -d "/opt/homebrew/Caskroom/flutter" ]]; then
    export PATH="/opt/homebrew/Caskroom/flutter/latest/flutter/bin:${PATH}"
  elif [[ -d "/usr/local/Caskroom/flutter" ]]; then
    export PATH="/usr/local/Caskroom/flutter/latest/flutter/bin:${PATH}"
  fi
fi

echo "==> Preparing Android SDK directories at ${SDK_ROOT} ..."
mkdir -p "${SDK_ROOT}/cmdline-tools" "${SDK_ROOT}/platform-tools"

echo "==> Downloading Android command-line tools ..."
TMP_DIR="$(mktemp -d)"
curl -fsSL "${CMDLINE_ZIP_URL}" -o "${TMP_DIR}/cmdline-tools.zip"
unzip -q "${TMP_DIR}/cmdline-tools.zip" -d "${TMP_DIR}"
rm -rf "${SDK_ROOT}/cmdline-tools/latest" || true
mkdir -p "${SDK_ROOT}/cmdline-tools/latest"
# Google zip extracts a 'cmdline-tools' folder; move its contents into 'latest'
cp -R "${TMP_DIR}/cmdline-tools/"* "${SDK_ROOT}/cmdline-tools/latest/"
rm -rf "${TMP_DIR}"

export ANDROID_SDK_ROOT="${SDK_ROOT}"
export ANDROID_HOME="${SDK_ROOT}"
export PATH="${SDK_ROOT}/cmdline-tools/latest/bin:${SDK_ROOT}/platform-tools:${PATH}"

echo "==> Accepting Android SDK licenses ..."
yes | sdkmanager --licenses >/dev/null || true

echo "==> Installing Android platform-tools, platforms and build-tools ..."
sdkmanager --install "platform-tools" >/dev/null || true
for p in "${PLATFORMS[@]}"; do
  sdkmanager --install "platforms;${p}" >/dev/null || true
done
for b in "${BUILD_TOOLS[@]}"; do
  sdkmanager --install "build-tools;${b}" >/dev/null || true
done

RC_FILE="$(detect_shell_rc)"
echo "==> Updating shell profile: ${RC_FILE}"
ensure_line_in_file "export ANDROID_SDK_ROOT=\"${SDK_ROOT}\"" "${RC_FILE}"
ensure_line_in_file "export ANDROID_HOME=\"${SDK_ROOT}\"" "${RC_FILE}"
ensure_line_in_file "export PATH=\"${SDK_ROOT}/cmdline-tools/latest/bin:\$PATH\"" "${RC_FILE}"
ensure_line_in_file "export PATH=\"${SDK_ROOT}/platform-tools:\$PATH\"" "${RC_FILE}"
# Add Flutter to PATH if not present
if ! command -v flutter >/dev/null 2>&1; then
  if [[ -d "/opt/homebrew/Caskroom/flutter" ]]; then
    ensure_line_in_file 'export PATH="/opt/homebrew/Caskroom/flutter/latest/flutter/bin:$PATH"' "${RC_FILE}"
  elif [[ -d "/usr/local/Caskroom/flutter" ]]; then
    ensure_line_in_file 'export PATH="/usr/local/Caskroom/flutter/latest/flutter/bin:$PATH"' "${RC_FILE}"
  fi
fi

echo "==> Running flutter doctor ..."
flutter doctor || true

echo "✅ macOS setup complete. Open a new shell or 'source' your RC file to refresh PATH."
