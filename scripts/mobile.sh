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

build_apk() {
  check_env || return 1
  python3 scripts/mobile-icons.py || return 1
  bash scripts/mobile-demo-data.sh || return 1
  local server_url="${J3_SERVER_URL:-}"
  if [ -n "$server_url" ]; then
    echo "📡 Baking venue URL: $server_url (offline demo fallback bundled)"
  else
    echo "📦 Building standalone offline demo (no venue URL baked)"
  fi
  VITE_API_BASE_URL="$server_url" VITE_OFFLINE_DEMO=1 VITE_QA_AUTOLOGIN=1 VITE_QA_FULL_PHONE=1 \
    bun run --cwd web build >/dev/null || return 1
  (cd web && bunx cap sync android) >/dev/null || return 1
  (cd web/android && ./gradlew assembleDebug --no-daemon) || return 1
  j3_prepare_build_dir "$J3_BUILD_ANDROID"
  cp "$APK" "$APK_OUT"
  cp "$APK" "$LOCAL_APK_OUT"
  cp web/static/j3-demo.db "$DB_OUT"
  cp web/static/j3-demo-data.json "$SNAPSHOT_OUT"
  echo "✅ APK ready: $LOCAL_APK_OUT (also $APK_OUT)"
  echo "   Bundled demo data: $DB_OUT and $SNAPSHOT_OUT"
}

build_ios() {
  python3 scripts/mobile-icons.py || return 1
  bash scripts/mobile-demo-data.sh || return 1
  local server_url="${J3_SERVER_URL:-}"
  if [ -n "$server_url" ]; then
    echo "📡 Baking venue URL into iOS client: $server_url (offline demo fallback bundled)"
  else
    echo "📦 Building standalone offline iOS demo (no venue URL baked)"
  fi
  VITE_API_BASE_URL="$server_url" VITE_OFFLINE_DEMO=1 VITE_QA_AUTOLOGIN=1 VITE_QA_FULL_PHONE=1 \
    bun run --cwd web build >/dev/null || return 1
  (cd web && bunx cap sync ios) >/dev/null || return 1

  j3_prepare_build_dir "$J3_BUILD_IOS"
  cp web/static/j3-demo.db "$J3_BUILD_IOS/j3-demo.db"
  cp web/static/j3-demo-data.json "$J3_BUILD_IOS/j3-demo-data.json"

  if command -v xcodebuild >/dev/null 2>&1; then
    local derived="$J3_BUILD_IOS/derivedData"
    echo "🍎 Building iOS simulator app → $J3_BUILD_IOS/J3.app"
    (cd web/ios/App && xcodebuild -project App.xcodeproj -scheme App \
      -configuration Release -sdk iphonesimulator \
      -derivedDataPath "$derived" CODE_SIGNING_ALLOWED=NO build) || return 1
    local app
    app=$(find "$derived/Build/Products" -type d -name App.app \
      -path '*Release-iphonesimulator*' 2>/dev/null | head -1)
    [ -n "$app" ] || { echo "❌ Xcode build produced no App.app"; return 1; }
    rm -rf "$J3_BUILD_IOS/J3.app"
    cp -R "$app" "$J3_BUILD_IOS/J3.app"
    echo "✅ iOS simulator app ready: $J3_BUILD_IOS/J3.app"
  else
    echo "ℹ️  Xcode not available here; Capacitor project synced at web/ios/App"
    echo "   Staged iOS demo assets in $J3_BUILD_IOS; build the .app on macOS with:"
    echo "   cd web/ios/App && xcodebuild -project App.xcodeproj -scheme App \\"
    echo "     -configuration Release -sdk iphonesimulator \\"
    echo "     -derivedDataPath ../../../build/ios/derivedData CODE_SIGNING_ALLOWED=NO build"
  fi
  echo "✅ iOS build assets ready: $J3_BUILD_IOS"
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
  build) build_apk ;;
  build-ios) build_ios ;;
  avd) ensure_avd ;;
  emulator) start_emulator ;;
  install) install_apk ;;
  url) lan_url ;;
  *)
    echo "Usage: scripts/mobile.sh {env|build|build-ios|avd|emulator|install|url}"
    exit 1
    ;;
esac
