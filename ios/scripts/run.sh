#!/usr/bin/env bash
# Builds Perchly and runs it in an iOS Simulator: regenerates the Xcode
# project, boots a simulator, builds for it, installs, and launches the app.
# Safe to re-run: every step is idempotent.
#
# Usage: ./scripts/run.sh ["iPhone 17 Pro"]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IOS_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

PROJECT="Perchly.xcodeproj"
SCHEME="Perchly"
BUNDLE_ID="com.perchly.app"
DEVICE_NAME="${1:-iPhone 17 Pro}"
DERIVED_DATA="$IOS_DIR/build"

log() { printf '\033[1;34m[run]\033[0m %s\n' "$1"; }
die() { printf '\033[1;31m[run]\033[0m %s\n' "$1" >&2; exit 1; }

command -v xcodebuild >/dev/null 2>&1 || die "xcodebuild bulunamadı. Xcode kurulu olmalı."
command -v xcrun >/dev/null 2>&1 || die "xcrun bulunamadı. Xcode command line tools kurulu olmalı."

cd "$IOS_DIR"

# 1. Regenerate the Xcode project from project.yml so newly added source
#    files are always picked up.
if command -v xcodegen >/dev/null 2>&1; then
  log "Xcode projesi project.yml'den yeniden üretiliyor..."
  xcodegen generate >/dev/null
else
  log "xcodegen bulunamadı, mevcut $PROJECT kullanılacak (yeni dosyalar eksik olabilir)."
fi

[ -d "$PROJECT" ] || die "$PROJECT bulunamadı."

# 2. Find a matching simulator (prefer one that's already booted).
UDID="$(
  xcrun simctl list devices available -j \
    | python3 -c "
import json, sys
data = json.load(sys.stdin)['devices']
name = '$DEVICE_NAME'
booted = None
first = None
for runtime, devices in data.items():
    if 'iOS' not in runtime:
        continue
    for d in devices:
        if d['name'] != name:
            continue
        if first is None:
            first = d['udid']
        if d['state'] == 'Booted':
            booted = d['udid']
if booted:
    print(booted)
elif first:
    print(first)
"
)"
[ -n "$UDID" ] || die "'$DEVICE_NAME' adında bir simülatör bulunamadı. 'xcrun simctl list devicetypes' ile mevcut adları görebilirsiniz."
log "Simülatör: $DEVICE_NAME ($UDID)"

# 3. Boot it if needed, and bring the Simulator app to the foreground.
STATE="$(xcrun simctl list devices available -j | python3 -c "
import json, sys
data = json.load(sys.stdin)['devices']
udid = '$UDID'
for devices in data.values():
    for d in devices:
        if d['udid'] == udid:
            print(d['state'])
")"
if [ "$STATE" != "Booted" ]; then
  log "Simülatör başlatılıyor..."
  xcrun simctl boot "$UDID"
fi
xcrun simctl bootstatus "$UDID" -b >/dev/null
open -a Simulator --args -CurrentDeviceUDID "$UDID"

# 4. Build for that simulator.
log "Proje derleniyor..."
xcodebuild \
  -project "$PROJECT" \
  -scheme "$SCHEME" \
  -destination "id=$UDID" \
  -configuration Debug \
  -derivedDataPath "$DERIVED_DATA" \
  build \
  | { grep -E "error:|BUILD SUCCEEDED|BUILD FAILED" || true; }

APP_PATH="$(find "$DERIVED_DATA/Build/Products" -maxdepth 2 -name "*.app" -path "*iphonesimulator*" | head -1)"
[ -n "$APP_PATH" ] || die "Derlenen .app bulunamadı."

# 5. Install and launch.
log "Uygulama kuruluyor..."
xcrun simctl install "$UDID" "$APP_PATH"

log "Uygulama başlatılıyor..."
xcrun simctl launch "$UDID" "$BUNDLE_ID" >/dev/null

log "Perchly simülatörde çalışıyor. (device: $DEVICE_NAME)"
