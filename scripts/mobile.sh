#!/usr/bin/env bash
# Mobile build, emulator and phone-QA helper for J3.
set -uo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
# shellcheck source=build-paths.sh
source scripts/build-paths.sh
j3_prepare_build_dirs

STUDIO_SDK="$HOME/Android/Sdk"
LEGACY_SDK="/home/linuxbrew/.linuxbrew/share/android-commandlinetools"
if [ -n "${ANDROID_HOME:-}" ]; then
  ANDROID_SDK="$ANDROID_HOME"
elif [ -d "$STUDIO_SDK" ]; then
  ANDROID_SDK="$STUDIO_SDK"
else
  ANDROID_SDK="$LEGACY_SDK"
fi
JAVA21="${JAVA_HOME:-/home/linuxbrew/.linuxbrew/Cellar/openjdk@21/21.0.13.1/libexec}"
AVD_NAME="${J3_AVD:-j3-pixel-10-pro}"
APK="web/android/app/build/outputs/apk/debug/app-debug.apk"
APK_OUT="$J3_BUILD_ANDROID/j3-demo.apk"
LOCAL_APK_OUT="$J3_BUILD_ANDROID/local.apk"
DEV_APK_OUT="$J3_BUILD_ANDROID/j3-dev.apk"
QA_APK_OUT="$J3_BUILD_ANDROID/j3-qa.apk"
PROD_APK_OUT="$J3_BUILD_ANDROID/j3-prod.apk"
DB_OUT="$J3_BUILD_ANDROID/j3-demo.db"
SNAPSHOT_OUT="$J3_BUILD_ANDROID/j3-demo-data.json"

setup_env() {
  export ANDROID_HOME="$ANDROID_SDK"
  export ANDROID_SDK_ROOT="$ANDROID_SDK"
  if [ -d "$JAVA21" ]; then
    export JAVA_HOME="$JAVA21"
    export PATH="$JAVA_HOME/bin:$PATH"
  fi
  export PATH="$ANDROID_SDK/emulator:$ANDROID_SDK/platform-tools:$PATH"
}

check_env() {
  setup_env
  command -v java >/dev/null || { echo "❌ Java not found"; return 1; }
  echo "Java: $(java -version 2>&1 | head -1)"
  echo "Android SDK: $ANDROID_HOME"
}

# Flavour rules shared by Android and iOS. `dev` targets the local venue server,
# `qa` is standalone with bundled seeded data, and `prod` requires a real public
# API origin. Only QA gets autologin and the full-phone route unlock.
set_flavor_env() {
  local flavor="$1"
  case "$flavor" in
    dev)
      FLAVOR_LABEL="dev-live"
      FLAVOR_SERVER_URL="${J3_DEV_API_URL:-$(lan_url)}"
      FLAVOR_OFFLINE=0
      FLAVOR_AUTOLOGIN=0
      FLAVOR_FULL_PHONE=0
      FLAVOR_BUNDLE_DATA=0
      ;;
    qa)
      FLAVOR_LABEL="qa-offline"
      FLAVOR_SERVER_URL=""
      FLAVOR_OFFLINE=1
      FLAVOR_AUTOLOGIN=1
      FLAVOR_FULL_PHONE=1
      FLAVOR_BUNDLE_DATA=1
      ;;
    prod)
      if [ -z "${J3_SERVER_URL:-}" ]; then
        echo "❌ J3_SERVER_URL is required for the production/live mobile build" >&2
        echo "   Example: J3_SERVER_URL=https://api.example.com bun run build:android:prod" >&2
        return 1
      fi
      FLAVOR_LABEL="prod-live"
      FLAVOR_SERVER_URL="$J3_SERVER_URL"
      FLAVOR_OFFLINE=0
      FLAVOR_AUTOLOGIN=0
      FLAVOR_FULL_PHONE=0
      FLAVOR_BUNDLE_DATA=0
      ;;
    *) echo "Unknown mobile flavour: $flavor" >&2; return 1 ;;
  esac
}

prepare_flavor() {
  local flavor="$1"
  set_flavor_env "$flavor" || return 1
  python3 scripts/mobile-icons.py || return 1
  if [ "$FLAVOR_BUNDLE_DATA" = "1" ]; then
    bash scripts/mobile-demo-data.sh || return 1
  fi
  if [ -n "$FLAVOR_SERVER_URL" ]; then
    echo "📡 Building $FLAVOR_LABEL against $FLAVOR_SERVER_URL"
  else
    echo "📦 Building $FLAVOR_LABEL with bundled seeded demo data"
  fi
}

build_web_bundle() {
  VITE_API_BASE_URL="$FLAVOR_SERVER_URL" \
  VITE_BUILD_FLAVOR="$FLAVOR_LABEL" \
  VITE_OFFLINE_DEMO="$FLAVOR_OFFLINE" \
  VITE_QA_AUTOLOGIN="$FLAVOR_AUTOLOGIN" \
  VITE_QA_FULL_PHONE="$FLAVOR_FULL_PHONE" \
    bun run --cwd web build >/dev/null
}

build_apk() {
  local flavor="${1:-qa}"
  local dest
  check_env || return 1
  prepare_flavor "$flavor" || return 1
  build_web_bundle || return 1
  (cd web && bunx cap sync android) >/dev/null || return 1
  (cd web/android && ./gradlew assembleDebug --no-daemon) || return 1

  j3_prepare_build_dir "$J3_BUILD_ANDROID"
  case "$flavor" in
    dev) dest="$DEV_APK_OUT" ;;
    qa) dest="$QA_APK_OUT" ;;
    prod) dest="$PROD_APK_OUT" ;;
    *) dest="$J3_BUILD_ANDROID/j3-$flavor.apk" ;;
  esac
  cp "$APK" "$dest"
  if [ "$flavor" = "qa" ]; then
    cp "$APK" "$APK_OUT"
    cp "$APK" "$LOCAL_APK_OUT"
    cp web/static/j3-demo.db "$DB_OUT"
    cp web/static/j3-demo-data.json "$SNAPSHOT_OUT"
    echo "   Bundled demo data: $DB_OUT and $SNAPSHOT_OUT"
  fi
  echo "✅ $FLAVOR_LABEL APK ready: $dest"
}

build_ios() {
  local flavor="${1:-qa}"
  local dest="$J3_BUILD_IOS"
  prepare_flavor "$flavor" || return 1
  build_web_bundle || return 1
  (cd web && bunx cap sync ios) >/dev/null || return 1

  if [ "$flavor" != "qa" ]; then
    dest="$J3_BUILD_IOS/$flavor"
  fi
  j3_prepare_build_dir "$dest"
  if [ "$FLAVOR_BUNDLE_DATA" = "1" ]; then
    cp web/static/j3-demo.db "$dest/j3-demo.db"
    cp web/static/j3-demo-data.json "$dest/j3-demo-data.json"
  fi

  if command -v xcodebuild >/dev/null 2>&1; then
    local derived="$dest/derivedData"
    echo "🍎 Building iOS simulator app → $dest/J3.app"
    (cd web/ios/App && xcodebuild -project App.xcodeproj -scheme App \
      -configuration Release -sdk iphonesimulator \
      -derivedDataPath "$derived" CODE_SIGNING_ALLOWED=NO build) || return 1
    local app
    app=$(find "$derived/Build/Products" -type d -name App.app \
      -path '*Release-iphonesimulator*' 2>/dev/null | head -1)
    [ -n "$app" ] || { echo "❌ Xcode build produced no App.app"; return 1; }
    rm -rf "$dest/J3.app"
    cp -R "$app" "$dest/J3.app"
    echo "✅ iOS simulator app ready: $dest/J3.app"
  else
    echo "ℹ️  Xcode not available here; Capacitor project synced at web/ios/App"
    echo "   Staged iOS assets in $dest; build the .app on macOS with:"
    echo "   cd web/ios/App && xcodebuild -project App.xcodeproj -scheme App \\"
    echo "     -configuration Release -sdk iphonesimulator \\"
    echo "     -derivedDataPath ../../../build/ios/derivedData CODE_SIGNING_ALLOWED=NO build"
  fi
  echo "✅ $FLAVOR_LABEL iOS assets ready: $dest"
}

ensure_avd() {
  check_env || return 1
  if avdmanager list avd 2>/dev/null | grep -q "Name: $AVD_NAME"; then
    echo "✅ AVD $AVD_NAME exists"
    return 0
  fi
  echo "📱 Creating Pixel 10 Pro AVD (Android 36)…"
  yes | sdkmanager --licenses >/dev/null 2>&1 || true
  sdkmanager "emulator" "platforms;android-36" "build-tools;36.0.0" "system-images;android-36;google_apis;x86_64" >/dev/null 2>&1
  echo no | avdmanager create avd -n "$AVD_NAME" -k 'system-images;android-36;google_apis;x86_64' -d pixel_10_pro --force >/dev/null 2>&1
  echo "✅ AVD created"
}

start_emulator() {
  ensure_avd || return 1
  if adb devices 2>/dev/null | grep -q 'emulator-.*device'; then
    echo "✅ Emulator already running"
    adb devices
    return 0
  fi
  echo "🚀 Starting ${AVD_NAME}…"
  nohup emulator -avd "$AVD_NAME" -no-snapshot -gpu auto -no-audio >/tmp/j3-emulator.log 2>&1 &
  adb wait-for-device
  echo "⏳ Waiting for boot (this can take up to two minutes)…"
  for _ in $(seq 1 90); do
    boot=$(adb shell getprop sys.boot_completed 2>/dev/null | tr -d '\r')
    [ "$boot" = "1" ] && break
    sleep 2
  done
  adb devices
}

install_apk() {
  check_env || return 1
  start_emulator || return 1
  if [ ! -f "$APK_OUT" ]; then
    echo "⚠️  No staged APK found — building first"
    build_apk || return 1
  fi
  echo "⬇️  Installing APK from $APK_OUT…"
  adb install -r "$APK_OUT" || return 1
  adb shell monkey -p com.j3.tournaments -c android.intent.category.LAUNCHER 1 >/dev/null 2>&1
  echo "✅ J3 launched on emulator"
}

lan_url() {
  local ip
  ip=$(ip -4 route get 8.8.8.8 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src") print $(i+1)}' | head -1)
  echo "http://${ip:-127.0.0.1}:8787"
}

case "${1:-}" in
  env) check_env ;;
  build|build-qa) build_apk qa ;;
  build-dev) build_apk dev ;;
  build-prod) build_apk prod ;;
  build-android-all)
    build_apk dev && build_apk qa && build_apk prod
    ;;
  build-ios|build-ios-qa) build_ios qa ;;
  build-ios-dev) build_ios dev ;;
  build-ios-prod) build_ios prod ;;
  build-ios-all)
    build_ios dev && build_ios qa && build_ios prod
    ;;
  avd) ensure_avd ;;
  emulator) start_emulator ;;
  install) install_apk ;;
  url) lan_url ;;
  *)
    echo "Usage: scripts/mobile.sh {env|build|build-dev|build-qa|build-prod|build-android-all|"
    echo "       build-ios|build-ios-dev|build-ios-qa|build-ios-prod|build-ios-all|"
    echo "       avd|emulator|install|url}"
    exit 1
    ;;
esac
