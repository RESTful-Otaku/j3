#!/usr/bin/env bash
# Install, launch and verify the J3 native client on a connected Android device or emulator.
# Used by CI and by local QA; all evidence is written to J3_MOBILE_SMOKE_OUT.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DEVICE="${J3_DEVICE:-}"
APK="${J3_APK:-web/android/app/build/outputs/apk/debug/app-debug.apk}"
OUT="${J3_MOBILE_SMOKE_OUT:-artifacts/mobile-smoke}"
PACKAGE="com.j3.tournaments"

mkdir -p "$OUT"

ADB=(adb)
if [ -n "$DEVICE" ]; then
  ADB+=(-s "$DEVICE")
fi

if ! command -v adb >/dev/null 2>&1; then
  echo "❌ adb not found on PATH"
  exit 1
fi
if [ ! -f "$APK" ]; then
  echo "❌ APK not found: $APK"
  exit 1
fi

timeout 180 "${ADB[@]}" wait-for-device
STATE="$("${ADB[@]}" get-state | tr -d '\r')"
if [ "$STATE" != "device" ]; then
  echo "❌ Android device is offline: $STATE"
  exit 1
fi

# Boot completion can precede a usable package manager by a few seconds.
for _ in $(seq 1 60); do
  if "${ADB[@]}" shell pm list packages >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

# Emulator images can reject side-loaded APKs while package verification is still
# settling; QA builds are local artefacts, so disable that check for the smoke run.
"${ADB[@]}" shell settings put global package_verifier_enable 0 >/dev/null 2>&1 || true
"${ADB[@]}" shell settings put global verify_apps_over_usb 0 >/dev/null 2>&1 || true

echo "📲 Installing $APK on $STATE"
INSTALL_LOG="$OUT/adb-install.txt"
INSTALLED=0
for attempt in 1 2 3; do
  if "${ADB[@]}" install -r "$APK" >"$INSTALL_LOG" 2>&1; then
    INSTALLED=1
    break
  fi
  echo "⚠️  Install attempt $attempt failed:"
  cat "$INSTALL_LOG"
  if grep -Eq 'INSTALL_FAILED_(UPDATE_INCOMPATIBLE|VERSION_DOWNGRADE|ALREADY_EXISTS|INVALID_APK|PERMISSIONS)' "$INSTALL_LOG"; then
    "${ADB[@]}" uninstall "$PACKAGE" >/dev/null 2>&1 || true
  fi
  sleep 3
done
if [ "$INSTALLED" -ne 1 ]; then
  echo "❌ APK installation failed (see $INSTALL_LOG)"
  "${ADB[@]}" logcat -d -t 500 > "$OUT/android-logcat.txt" || true
  exit 1
fi
"${ADB[@]}" shell am force-stop "$PACKAGE" >/dev/null 2>&1 || true

# Launch the Capacitor activity directly; monkey is a fallback for shells where
# `am start` is unavailable. Keep output so a launch failure is diagnosable.
LAUNCH_LOG="$OUT/adb-launch.txt"
if ! "${ADB[@]}" shell am start -n "$PACKAGE/com.j3.tournaments.MainActivity" >"$LAUNCH_LOG" 2>&1; then
  echo "⚠️  am start failed; trying monkey launcher"
  cat "$LAUNCH_LOG"
  "${ADB[@]}" shell monkey -p "$PACKAGE" -c android.intent.category.LAUNCHER 1 >>"$LAUNCH_LOG" 2>&1 || true
fi

PID=""
for _ in $(seq 1 60); do
  # `pidof` exits 1 until the process exists; with `pipefail` that would abort
  # the whole smoke script before the retry loop gets a chance.
  PID="$("${ADB[@]}" shell pidof "$PACKAGE" 2>/dev/null | tr -d '\r' | awk '{print $1}' || true)"
  [ -n "$PID" ] && break
  sleep 1
done
if [ -z "$PID" ]; then
  echo "❌ $PACKAGE did not start"
  "${ADB[@]}" logcat -d -t 500 > "$OUT/android-logcat.txt" || true
  exit 1
fi

# Give the WebView a moment to render the bundled SPA before capturing evidence.
sleep 6

"${ADB[@]}" shell uiautomator dump /sdcard/j3-window.xml >/dev/null 2>&1 || true
"${ADB[@]}" pull /sdcard/j3-window.xml "$OUT/window.xml" >/dev/null 2>&1 || true
if [ ! -s "$OUT/window.xml" ] || ! grep -Eqi 'J3|J Cubed|Public|Okami|Tournament' "$OUT/window.xml"; then
  echo "❌ Native client UI did not render recognisable J3 content"
  exit 1
fi

"${ADB[@]}" exec-out screencap -p > "$OUT/android-running.png"
if [ ! -s "$OUT/android-running.png" ]; then
  echo "❌ Screenshot capture failed"
  exit 1
fi

"${ADB[@]}" logcat -d -t 1000 > "$OUT/android-logcat.txt"
if grep -Ei 'FATAL EXCEPTION|AndroidRuntime.*'"$PACKAGE"'|java\.lang\.(RuntimeException|IllegalStateException|NullPointerException)' "$OUT/android-logcat.txt"; then
  echo "❌ Native client crashed"
  exit 1
fi

echo "✅ Native client launched, rendered and captured (pid $PID)"
echo "   evidence: $OUT"
