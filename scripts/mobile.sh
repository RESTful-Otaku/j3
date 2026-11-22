#!/usr/bin/env bash
# Mobile build, emulator and phone-QA helper for J3.
set -uo pipefail
export PATH="$HOME/.bun/bin:$PATH"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

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
  local server_url="${J3_SERVER_URL:-}"
  if [ -n "$server_url" ]; then
    echo "📡 Baking venue URL: $server_url"
  fi
  VITE_API_BASE_URL="$server_url" bun run --cwd web build >/dev/null || return 1
  (cd web && bunx cap sync android) >/dev/null || return 1
  (cd web/android && ./gradlew assembleDebug --no-daemon) || return 1
  mkdir -p artifacts
  cp "$APK" artifacts/j3-demo.apk
  echo "✅ APK ready: artifacts/j3-demo.apk"
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
  echo "🚀 Starting $AVD_NAME…"
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
  if [ ! -f "$APK" ]; then
    echo "⚠️  No APK found — building first"
    build_apk || return 1
  fi
  echo "⬇️  Installing APK…"
  adb install -r "$APK" || return 1
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
  avd) ensure_avd ;;
  emulator) start_emulator ;;
  install) install_apk ;;
  url) lan_url ;;
  *)
    echo "Usage: scripts/mobile.sh {env|build|avd|emulator|install|url}"
    exit 1
    ;;
esac
