#Requires -Version 5.1
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# Windows Flutter + Android SDK bootstrap (PowerShell)
# - Installs Flutter (git clone stable) and Android cmdline tools (no full Android Studio)
# - Sets persistent user env vars
# - Installs Android API 36/35 platforms and build-tools
# - Accepts licenses and runs flutter doctor
#
# Run PowerShell as the current user (Admin not strictly required unless your environment restricts env updates)

# Config
$SdkRoot = Join-Path $env:LOCALAPPDATA "Android\sdk"
$CmdlineZipUrl = "https://dl.google.com/android/repository/commandlinetools-win-11076708_latest.zip"
$FlutterDir = Join-Path $env:USERPROFILE "flutter"
$BuildTools = @("36.0.0", "35.0.0")
$Platforms = @("android-36", "android-35")

function Ensure-Dir($path) { if (-not (Test-Path $path)) { New-Item -Path $path -ItemType Directory | Out-Null } }

function Add-UserPathIfMissing($dir) {
  $current = [Environment]::GetEnvironmentVariable("Path", "User")
  if (-not ($current -split ';' | Where-Object { $_ -eq $dir })) {
    [Environment]::SetEnvironmentVariable("Path", ($current + ";" + $dir), "User")
    Write-Host "Added to user PATH: $dir"
  }
}

Write-Host "==> Ensuring Git (required for Flutter clone) ..."
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
  Write-Warning "Git not found. Please install Git for Windows from https://git-scm.com/download/win and re-run this script."
  exit 1
}

Write-Host "==> Installing Flutter (stable) to $FlutterDir ..."
if (-not (Test-Path $FlutterDir)) {
  git clone https://github.com/flutter/flutter.git -b stable $FlutterDir
} else {
  Push-Location $FlutterDir
  git fetch; git checkout stable; git pull
  Pop-Location
}

Write-Host "==> Preparing Android SDK directories at $SdkRoot ..."
Ensure-Dir $SdkRoot
Ensure-Dir (Join-Path $SdkRoot "cmdline-tools")
Ensure-Dir (Join-Path $SdkRoot "platform-tools")

$Tmp = New-Item -ItemType Directory -Path ([System.IO.Path]::GetTempPath()) -Name ("cmdlinetools_" + [System.Guid]::NewGuid().ToString()) -Force
$ZipPath = Join-Path $Tmp "cmdline-tools.zip"

Write-Host "==> Downloading Android command-line tools ..."
Invoke-WebRequest -Uri $CmdlineZipUrl -OutFile $ZipPath
Expand-Archive -Path $ZipPath -DestinationPath $Tmp -Force
$LatestDir = Join-Path $SdkRoot "cmdline-tools\latest"
if (Test-Path $LatestDir) { Remove-Item $LatestDir -Recurse -Force }
New-Item -ItemType Directory -Path $LatestDir | Out-Null
Copy-Item -Recurse -Force (Join-Path $Tmp "cmdline-tools\*") $LatestDir
Remove-Item $Tmp -Recurse -Force

# Persistent env vars
[Environment]::SetEnvironmentVariable("ANDROID_SDK_ROOT", $SdkRoot, "User")
[Environment]::SetEnvironmentVariable("ANDROID_HOME", $SdkRoot, "User")
Add-UserPathIfMissing (Join-Path $FlutterDir "bin")
Add-UserPathIfMissing (Join-Path $SdkRoot "cmdline-tools\latest\bin")
Add-UserPathIfMissing (Join-Path $SdkRoot "platform-tools")

# Session env
$env:ANDROID_SDK_ROOT = $SdkRoot
$env:ANDROID_HOME = $SdkRoot
$env:Path = (Join-Path $FlutterDir "bin") + ";" + (Join-Path $SdkRoot "cmdline-tools\latest\bin") + ";" + (Join-Path $SdkRoot "platform-tools") + ";" + $env:Path

Write-Host "==> Accepting Android SDK licenses ..."
cmd /c "echo y | sdkmanager --licenses" | Out-Null

Write-Host "==> Installing Android platform-tools, platforms and build-tools ..."
sdkmanager --install "platform-tools" | Out-Null
foreach ($p in $Platforms) { sdkmanager --install ("platforms;" + $p) | Out-Null }
foreach ($b in $BuildTools) { sdkmanager --install ("build-tools;" + $b) | Out-Null }

Write-Host "==> Running flutter doctor ..."
flutter doctor

Write-Host "✅ Windows setup complete. You may need to open a new terminal for PATH changes to take effect."
